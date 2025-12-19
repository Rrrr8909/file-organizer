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
	sourceDir          string
	rules              map[string]string
	processedFiles     int
	processedFilesSize int64
	logFile            *os.File
	statistics         map[string]*FileStats
}

func NewFileOrganizer(sourceDir string) *FileOrganizer {
	file, err := os.OpenFile("organizer.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	log.SetOutput(file)

	return &FileOrganizer{
		sourceDir:          sourceDir,
		rules:              DefaultRules,
		processedFiles:     0,
		processedFilesSize: 0,
		logFile:            file,
		statistics:         make(map[string]*FileStats),
	}
}

func (fo *FileOrganizer) Close() error {
	if fo.logFile != nil {
		return fo.logFile.Close()
	}
	return nil
}

func (fo *FileOrganizer) logSuccess(message string) {
	log.Printf("%s [SUCCESS] %s", time.Now().Format("2006/01/02 15:04:05"), message)
}

func (fo *FileOrganizer) logError(message string) {
	log.Printf("%s [ERROR] %s", time.Now().Format("2006/01/02 15:04:05"), message)
}

func (fo *FileOrganizer) moveFile(sourcePath, targetDir string) error {
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return err
	}

	targetPath := filepath.Join(targetDir, filepath.Base(sourcePath))

	err := os.Rename(sourcePath, targetPath)
	if err != nil {
		return err
	}

	return nil
}

func (fo *FileOrganizer) Organize() error {
	err := filepath.Walk(fo.sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(info.Name())

		targetDir, ok := fo.rules[ext]
		if !ok {
			fo.logError(fmt.Sprintf("%s: extension %s not supported", path, ext))
			return nil
		}

		err = fo.moveFile(path, targetDir)
		if err != nil {
			fo.logError(fmt.Sprintf("move error %s -> %s: %v", path, targetDir, err))
			return err
		}

		fs, ok := fo.statistics[targetDir]
		if !ok {
			fs = &FileStats{}
			fo.statistics[targetDir] = fs
		}
		fs.Count++
		fs.TotalSize += info.Size()

		fo.processedFiles++
		fo.processedFilesSize += info.Size()

		fo.logSuccess(fmt.Sprintf("moved: %s -> %s", path, targetDir))

		return nil
	})

	if err != nil {
		fo.logError(fmt.Sprintf("error reading source directory: %s", err))
		return err
	}

	return nil
}

func (fo *FileOrganizer) Report() {
	fmt.Printf("=== Отчет о перемещении файлов === \n\n")
	fmt.Printf("Всего обработано файлов: %d \n", fo.processedFiles)
	fmt.Printf("Общий размерЖ: %d B \n\n", fo.processedFilesSize)

	fmt.Println("Статистика по категориям: ")
	for k, v := range fo.statistics {
		fmt.Printf("%s: \n - Количество файлов: %1d \n - Общий размер: %d B \n", k, v.Count, v.TotalSize)
	}
}
func getUserInput(r *bufio.Reader) string {
	line, err := r.ReadString('\n')
	if err != nil {
		if err == io.EOF && strings.TrimSpace(line) == "" {
			return ""
		}
		panic(err)
	}

	return strings.TrimSpace(line)
}

func main() {
	in := bufio.NewReader(os.Stdin)

	fmt.Println("Программа организации структуры файлов")
	for {
		fmt.Println("Введите путь к неструктурированной директории: ")

		dir := getUserInput(in)
		if dir == "" {
			fmt.Println("Конец программы")
			break
		}
		fo := NewFileOrganizer(dir)

		fmt.Println("Начинаем структурировать файлы по папкам...")
		err := fo.Organize()
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("Конец работы")
		fo.Report()

		err = fo.Close()
		if err != nil {
			panic(err)
		}
	}
}
