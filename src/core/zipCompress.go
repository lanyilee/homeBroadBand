package core

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

//压缩文件
//files 文件数组，可以是不同dir下的文件或者文件夹
//dest 压缩文件存放地址
func Compress(files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()
	for _, file := range files {
		err := CompressSingle(file, "", w)
		if err != nil {
			return err
		}
	}
	return nil
}

//prefix 是解压后的目录
func CompressSingle(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = CompressSingle(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

//解压
func DeCompress_zip(zipFile, dest_path string) error {
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer reader.Close()
	DeCompress_to_dest := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()
		filename := dest_path + f.Name
		if err = os.MkdirAll(filepath.Dir(filename), 0755); err != nil {
			return err
		}
		w, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer w.Close()
		_, err = io.Copy(w, rc)
		return err
	}
	for _, file := range reader.File {
		err := DeCompress_to_dest(file)
		if err != nil {
			return err
		}
	}
	return nil
}

