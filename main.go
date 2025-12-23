package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var DefaultRules = map[string]string{
	".jpg":  "Images",
	".jpeg": "Images",
	".png":  "Images",
	".pdf":  "Documents",
	".doc":  "Documents",
	".docx": "Documents",
	".txt":  "Documents",
	".mp3":  "Music",
	".wav":  "Music",
	".mp4":  "Video",
	".avi":  "Video",
	".zip":  "Archives",
	".rar":  "Archives",
}

type FileStats struct {
	Count     int
	TotalSize int64
}

type FileOrganizer struct {
	sourceDir      string
	rules          map[string]string
	processedFiles int
	logFile        *os.File
	logger         *log.Logger
	statistics     map[string]*FileStats
}

func NewFileOrganizer(sourceDir string) (*FileOrganizer, error) {
	file, err := os.OpenFile("organizer.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileOrganizer{
		sourceDir:      sourceDir,
		rules:          DefaultRules,
		processedFiles: 0,
		logFile:        file,
		logger:         log.New(file, "", log.LstdFlags),
		statistics:     make(map[string]*FileStats),
	}, nil
}

func (fo *FileOrganizer) Close() error {
	if fo.logFile != nil {
		return fo.logFile.Close()
	}
	return nil
}

func (fo *FileOrganizer) logSuccess(message string) {
	fo.logger.Printf("[SUCCESS] %s", message)
}

func (fo *FileOrganizer) logError(message string) {
	fo.logger.Printf("[ERROR] %s", message)
}

func (fo *FileOrganizer) moveFile(sourcePath, targetDir string) error {
	targetDir = filepath.Join(fo.sourceDir, targetDir)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	baseName := filepath.Base(sourcePath)
	ext := filepath.Ext(baseName)

	targetPath := filepath.Join(targetDir, baseName)

	if _, err := os.Stat(targetPath); err == nil {
		baseName = fmt.Sprintf("%s_%s%s", baseName[:len(baseName)-len(ext)], time.Now().Format("2006-01-02_15-04-05"), ext)
		targetPath = filepath.Join(targetDir, baseName)
	} else if !os.IsNotExist(err) {
		return err
	}

	err := os.Rename(sourcePath, targetPath)
	if err != nil {
		return err
	}

	return nil
}

func (fo *FileOrganizer) Organize() error {

	files, err := os.ReadDir(fo.sourceDir)
	if err != nil {
		fo.logError(fmt.Sprintf("error reading source directory: %s", err))
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		info, err := file.Info()
		if err != nil {
			fo.logError(fmt.Sprintf("error reading file: %s", err))
			continue
		}

		name := file.Name()
		ext := filepath.Ext(name)

		path := filepath.Join(fo.sourceDir, name)

		targetDir, ok := fo.rules[ext]
		if !ok {
			fo.logError(fmt.Sprintf("%s: extension %s not supported", path, ext))
			continue
		}

		err = fo.moveFile(path, targetDir)
		if err != nil {
			fo.logError(fmt.Sprintf("move error %s -> %s: %v", path, targetDir, err))
			continue
		}

		fs, ok := fo.statistics[targetDir]
		if !ok {
			fs = &FileStats{}
			fo.statistics[targetDir] = fs
		}
		fs.Count++
		fs.TotalSize += info.Size()

		fo.processedFiles++

		fo.logSuccess(fmt.Sprintf("moved: %s -> %s", path, targetDir))
	}

	return nil
}

func (fo *FileOrganizer) Report() {
	fmt.Printf("=== Отчет о перемещении файлов === \n\n")
	fmt.Printf("Всего обработано файлов: %d \n", fo.processedFiles)
	var totalSize int64
	for _, v := range fo.statistics {
		totalSize += v.TotalSize
	}
	fmt.Printf("Общий размер: %.1f MB \n\n", float64(totalSize)/1024/1024)

	fmt.Println("Статистика по категориям: ")
	for k, v := range fo.statistics {
		fmt.Printf("%s: \n - Количество файлов: %d \n - Общий размер: %.1f MB \n", k, v.Count, float64(v.TotalSize)/1024/1024)
	}
}
func getUserInput(r *bufio.Reader) (string, error) {
	line, err := r.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			line = strings.TrimSpace(line)
			if line == "" {
				return "", io.EOF
			}
			return line, nil
		}
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func main() {
	in := bufio.NewReader(os.Stdin)

	fmt.Println("Программа организации структуры файлов")
	for {
		fmt.Println("Введите путь к неструктурированной директории: ")

		dir, err := getUserInput(in)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println("Ошибка ввода: ", err)
			continue
		}
		if dir == "" {
			break
		}
		fo, err := NewFileOrganizer(dir)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Начинаем структурировать файлы по папкам...")
		err = fo.Organize()
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Конец работы")
			fo.Report()
		}

		err = fo.Close()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println()
	}

	fmt.Println("Конец программы")
}
