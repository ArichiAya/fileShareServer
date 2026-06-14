package v1

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type FSserver struct {
	uploadDir string
	trashDir  string
}

type Option func(s *FSserver)

const maxUploadSize = 10_000_000_000 // 10 GB

func NewFSserver(options ...Option) *FSserver {
	srv := &FSserver{
		uploadDir: "G:/fileShareServer/upload",
		trashDir:  "G:/fileShareServer/trash",
	}

	for _, option := range options {
		option(srv)
	}

	return srv
}

func WithUploadDir(uploadDir string) Option {
	return func(s *FSserver) {
		s.uploadDir = uploadDir
	}
}

func WithTrashDir(trashDir string) Option {
	return func(s *FSserver) {
		s.trashDir = trashDir
	}
}

// safeJoin 将 dir 安全拼接在 base 后，防止路径穿越
func safeJoin(base, dir string) (string, error) {
	cleanBase := filepath.Clean(base)
	var target string
	if dir == "" {
		target = cleanBase
	} else {
		target = filepath.Join(cleanBase, filepath.Clean(dir))
	}
	prefix := cleanBase + string(filepath.Separator)
	if target != cleanBase && !strings.HasPrefix(target, prefix) {
		return "", fmt.Errorf("非法路径")
	}
	return target, nil
}

// moveSafe 安全移动 srcPath 到 dstPath。
// 如果目标已存在（同名的文件或目录），会合并内容而非覆盖。
func moveSafe(srcPath, dstPath string) error {
	// 目标不存在：简单 rename
	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		return os.Rename(srcPath, dstPath)
	}

	// 目标存在
	srcInfo, srcErr := os.Stat(srcPath)
	dstInfo, dstErr := os.Stat(dstPath)
	if srcErr != nil || dstErr != nil {
		return fmt.Errorf("无法访问源或目标")
	}

	// 文件 → 文件冲突
	if !srcInfo.IsDir() || !dstInfo.IsDir() {
		return fmt.Errorf("目标已存在: %s", dstPath)
	}

	// 都是目录：逐项合并
	entries, err := os.ReadDir(srcPath)
	if err != nil {
		return err
	}
	var errs []error
	for _, entry := range entries {
		srcItem := filepath.Join(srcPath, entry.Name())
		dstItem := filepath.Join(dstPath, entry.Name())
		if err := moveSafe(srcItem, dstItem); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("部分移动失败: %v", errs)
	}
	// 全部移空后删除源目录
	return os.Remove(srcPath)
}

func (s *FSserver) Upload(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	// 第一步：限制请求体大小，防止撑爆内存/磁盘
	req.Body = http.MaxBytesReader(w, req.Body, maxUploadSize)

	// 第二步：解析 multipart 表单（内存中最多缓存 10MiB，超过落磁盘临时文件）
	if err := req.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "文件太大或格式错误", http.StatusBadRequest)
		return
	}
	defer req.MultipartForm.RemoveAll() // 清理临时文件

	// 第三步：拿到文件
	file, header, err := req.FormFile("file") // "file" 对应前端表单的字段名
	if err != nil {
		http.Error(w, "未找到文件字段", http.StatusBadRequest)
		return
	}
	defer file.Close()

	yearMonth := time.Now().Format("2006-01") // Go 的魔数时间：2026-06
	dstDir := filepath.Join(s.uploadDir, yearMonth)
	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		http.Error(w, "创建目录失败", http.StatusInternalServerError)
		return
	}

	// 第四步：创建目标文件（Base 防路径穿越）
	safeName := filepath.Base(header.Filename)
	dstPath := filepath.Join(dstDir, safeName)
	dst, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "创建文件失败", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// 第五步：边拷贝边检查 context 是否被取消
	done := make(chan error, 1)
	var written int64

	go func() {
		n, err := io.Copy(dst, file)
		written = n
		done <- err
	}()

	select {
	case <-ctx.Done():
		// 客户端断开了，删掉半截文件
		// 先关再删（Windows 要求），defer 再关一次无害
		dst.Close()
		os.Remove(dstPath)
		return
	case err := <-done:
		if err != nil {
			os.Remove(dstPath)
			http.Error(w, "写入失败", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "上传成功: %s (%d bytes)", header.Filename, written)
}

func (s *FSserver) Download(w http.ResponseWriter, req *http.Request) {
	// 第一步：从查询参数获取目录和文件名
	// 示例: GET /download?dir=2026-06&name=photo.png
	dir := req.URL.Query().Get("dir")
	name := req.URL.Query().Get("name")

	if dir == "" || name == "" {
		http.Error(w, "缺少 dir 或 name 参数", http.StatusBadRequest)
		return
	}

	// 第二步：拼接绝对路径，防止路径穿越
	safeName := filepath.Base(name)
	dirPath, err := safeJoin(s.uploadDir, dir)
	if err != nil {
		http.Error(w, "非法路径", http.StatusForbidden)
		return
	}
	filePath := filepath.Join(dirPath, safeName)

	// 第四步：检查文件是否存在，且不是目录
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "文件不存在", http.StatusNotFound)
		} else {
			http.Error(w, "打开文件失败", http.StatusInternalServerError)
		}
		return
	}
	if info.IsDir() {
		http.Error(w, "不能下载目录", http.StatusBadRequest)
		return
	}

	// 第五步：设置下载响应头，交给 ServeFile 处理（自动支持 Range、Content-Type）
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, safeName))
	http.ServeFile(w, req, filePath)
}

func (s *FSserver) ListFiles(w http.ResponseWriter, req *http.Request) {
	// 第一步：获取参数
	// 示例: GET /listfiles?dir=2026-06
	// trash=1 时列出 trashDir 下的文件
	dir := req.URL.Query().Get("dir")
	trash := req.URL.Query().Get("trash")

	// 第二步：选择基础目录
	baseDir := s.uploadDir
	if trash == "1" {
		baseDir = s.trashDir
	}

	dirPath, err := safeJoin(baseDir, dir)
	if err != nil {
		http.Error(w, "非法路径", http.StatusForbidden)
		return
	}

	// 第三步：检查目录是否存在
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "文件夹不存在", http.StatusNotFound)
		} else {
			http.Error(w, "打开文件夹失败", http.StatusInternalServerError)
		}
		return
	}
	if !info.IsDir() {
		http.Error(w, "不是文件夹", http.StatusBadRequest)
		return
	}

	// 第四步：读取目录内容
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		http.Error(w, "读取文件夹失败", http.StatusInternalServerError)
		return
	}

	// 第五步：组装文件信息列表
	type FileInfo struct {
		Name  string    `json:"name"`
		Size  int64     `json:"size"`
		IsDir bool      `json:"isDir"`
		Time  time.Time `json:"time"`
	}

	files := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info() // 获取 os.FileInfo 拿 Size
		if err != nil {
			continue // 个别文件获取失败则跳过
		}
		files = append(files, FileInfo{
			Name:  entry.Name(),
			Size:  info.Size(),
			IsDir: entry.IsDir(),
			Time:  info.ModTime(),
		})
	}

	// 第六步：按名称降序排列（年月目录天然从新到旧）
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name > files[j].Name
	})

	// 第七步：返回 JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func (s *FSserver) Delete(w http.ResponseWriter, req *http.Request) {
	// 第一步：从查询参数获取目录和文件名
	// 示例: DELETE /delete?dir=2026-06&name=photo.png
	// 删文件夹时 name 不传: DELETE /delete?dir=2026-06
	dir := req.URL.Query().Get("dir")
	name := req.URL.Query().Get("name")

	if dir == "" {
		http.Error(w, "缺少 dir 参数", http.StatusBadRequest)
		return
	}

	// 第一步：构建源路径
	dirPath, err := safeJoin(s.uploadDir, dir)
	if err != nil {
		http.Error(w, "非法路径", http.StatusForbidden)
		return
	}

	var srcPath string
	if name == "" {
		srcPath = dirPath
	} else {
		srcPath = filepath.Join(dirPath, filepath.Base(name))
	}

	// 第二步：检查源是否存在
	if _, err := os.Stat(srcPath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "文件或文件夹不存在", http.StatusNotFound)
		} else {
			http.Error(w, "访问失败", http.StatusInternalServerError)
		}
		return
	}

	// 第三步：在 trash 中保持相同目录结构
	trashDirPath, err := safeJoin(s.trashDir, dir)
	if err != nil {
		http.Error(w, "非法路径", http.StatusForbidden)
		return
	}

	var dstPath string
	if name == "" {
		dstPath = trashDirPath
	} else {
		dstPath = filepath.Join(trashDirPath, filepath.Base(name))
		if err := os.MkdirAll(trashDirPath, 0o755); err != nil {
			http.Error(w, "创建回收目录失败", http.StatusInternalServerError)
			return
		}
	}

	// 第四步：移动文件/文件夹（目标已存在时合并目录内容）
	if err := moveSafe(srcPath, dstPath); err != nil {
		http.Error(w, fmt.Sprintf("删除失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "已移至回收站: %s", srcPath)
}

func (s *FSserver) Restore(w http.ResponseWriter, req *http.Request) {
	// 第一步：从查询参数获取目录和文件名
	// 示例: POST /restore?dir=2026-06&name=photo.png
	// 恢复文件夹时 name 不传: POST /restore?dir=2026-06
	dir := req.URL.Query().Get("dir")
	name := req.URL.Query().Get("name")

	if dir == "" {
		http.Error(w, "缺少 dir 参数", http.StatusBadRequest)
		return
	}

	// 第一步：构建 trash 中的源路径
	trashDirPath, err := safeJoin(s.trashDir, dir)
	if err != nil {
		http.Error(w, "非法路径", http.StatusForbidden)
		return
	}

	var srcPath string
	if name == "" {
		srcPath = trashDirPath
	} else {
		srcPath = filepath.Join(trashDirPath, filepath.Base(name))
	}

	// 第二步：检查源是否存在
	if _, err := os.Stat(srcPath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "trash 中文件或文件夹不存在", http.StatusNotFound)
		} else {
			http.Error(w, "访问失败", http.StatusInternalServerError)
		}
		return
	}

	// 第三步：构建目标路径
	dstDirPath, err := safeJoin(s.uploadDir, dir)
	if err != nil {
		http.Error(w, "非法路径", http.StatusForbidden)
		return
	}

	var dstPath string
	if name == "" {
		dstPath = dstDirPath
	} else {
		dstPath = filepath.Join(dstDirPath, filepath.Base(name))
		if err := os.MkdirAll(dstDirPath, 0o755); err != nil {
			http.Error(w, "创建恢复目录失败", http.StatusInternalServerError)
			return
		}
	}

	// 第四步：执行恢复（目标已存在时合并目录内容，文件冲突则报错）
	if err := moveSafe(srcPath, dstPath); err != nil {
		http.Error(w, fmt.Sprintf("恢复失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"success": true,
		"path":    dstPath,
	})
}

func (s *FSserver) DownloadDir(w http.ResponseWriter, req *http.Request) {
	// 示例: GET /downloaddir?dir=2026-06
	dir := req.URL.Query().Get("dir")
	if dir == "" {
		http.Error(w, "缺少 dir 参数", http.StatusBadRequest)
		return
	}

	// 构建安全路径
	dirPath, err := safeJoin(s.uploadDir, dir)
	if err != nil {
		http.Error(w, "非法路径", http.StatusForbidden)
		return
	}

	// 检查存在且是目录
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "目录不存在", http.StatusNotFound)
		} else {
			http.Error(w, "访问失败", http.StatusInternalServerError)
		}
		return
	}
	if !info.IsDir() {
		http.Error(w, "不是目录", http.StatusBadRequest)
		return
	}

	// 设置响应头，流式生成 zip
	dirName := filepath.Base(dirPath)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, dirName))

	zw := zip.NewWriter(w)
	defer zw.Close()

	filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == dirPath {
			return nil // 跳过根目录本身
		}

		relPath, _ := filepath.Rel(dirPath, path)

		if d.IsDir() {
			// 在 zip 中创建目录条目
			_, err := zw.CreateHeader(&zip.FileHeader{
				Name:     relPath + string(filepath.Separator),
				Modified: time.Now(),
			})
			return err
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		fInfo, err := f.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(fInfo)
		if err != nil {
			return err
		}
		header.Name = relPath
		header.Method = zip.Deflate

		entryW, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(entryW, f)
		return err
	})
}

func (s *FSserver) Clean(w http.ResponseWriter, req *http.Request) {
	// 第一步：从查询参数获取目录和文件名
	// 示例: DELETE /clean?dir=2026-06&name=photo.png
	// 清空文件夹时 name 不传: DELETE /clean?dir=2026-06
	dir := req.URL.Query().Get("dir")
	name := req.URL.Query().Get("name")

	if dir == "" {
		http.Error(w, "缺少 dir 参数", http.StatusBadRequest)
		return
	}

	// 第一步：构建 trash 中的路径
	trashDirPath, err := safeJoin(s.trashDir, dir)
	if err != nil {
		http.Error(w, "非法路径", http.StatusForbidden)
		return
	}

	var srcPath string
	if name == "" {
		srcPath = trashDirPath
	} else {
		srcPath = filepath.Join(trashDirPath, filepath.Base(name))
	}

	// 第二步：检查源是否存在
	if _, err := os.Stat(srcPath); err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "文件或文件夹不存在", http.StatusNotFound)
		} else {
			http.Error(w, "访问失败", http.StatusInternalServerError)
		}
		return
	}

	// 第三步 执行删除
	if err := os.RemoveAll(srcPath); err != nil {
		http.Error(w, fmt.Sprintf("删除失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "删除成功: %s", srcPath)
}
