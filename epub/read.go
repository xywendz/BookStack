package epub

import (
	"errors"
	"io/ioutil"
	"strings"
)

// 返回一个章节的内容
func (p *Book) ReadChapter(n string) (string, error) {
	if p == nil {
		return "", errors.New("nil pointer receiver")
	}

	// 如果 n 是一个标题或文件链接，则打开对应的章节内容
	for _, point := range p.Ncx.Points {
		if point.Text == n || strings.Contains(point.Content.Src, n) {
			src := point.Content.Src
			if strings.Contains(src, "#") {
				parts := strings.Split(src, "#")
				if len(parts) != 2 {
					return "", errors.New("路径不止一个锚点," + src)
				}
				src = parts[0]
			}
			content, _ := p.readFile(src)
			return content, nil
		}
	}

	// 如果 n 既不是一个文件名也不是一个标题，则返回错误
	return "", errors.New("找不到该文件," + n)
}

// 返回一个文件的内容
func (p *Book) readFile(n string) (string, error) {
	src := p.filename(n)
	fd, err := p.open(src)
	if err != nil {
		return "", err
	}
	defer fd.Close()
	b, err := ioutil.ReadAll(fd)
	if err != nil {
		return "", err
	}

	// 返回整个文件内容
	return string(b), nil
}

// 返回所有章节的内容
func (p *Book) ReadAll() ([][]string, error) {
	if p == nil {
		return nil, errors.New("nil pointer receiver")
	}
	tocMap := make(map[string]string)
	for _, point := range p.Ncx.Points {
		pp := point.Points
		ch := point.Text
		if len(pp) > 0 {
			for _, p := range pp {
				tocMap[p.Content.Src] = ch + "_" + p.Text
			}
		} else {
			tocMap[point.Content.Src] = ch
		}
	}

	itemMap := make(map[string]string)
	for _, manifest := range p.Opf.Manifest {
		itemMap[manifest.ID] = manifest.Href
	}

	var results [][]string
	var readAll func(spineItem []SpineItem) error
	readAll = func(spineItem []SpineItem) error {
		var ch string
		var content string
		for _, item := range spineItem {
			id := item.IDref
			src := itemMap[id]
			data, err := p.readFile(src)
			if err != nil {
				return err
			}
			if ch == "" { //第一章节
				ch = tocMap[src]
			}
			var curCh string
			if tocMap[src] == "" {
				curCh = ch
			} else {
				curCh = tocMap[src]
			}

			if curCh != ch { //新的章节
				var chData []string
				chData = append(chData, ch)
				chData = append(chData, content)
				results = append(results, chData)
				content = data
				ch = tocMap[src]
			} else {
				content = content + data
			}

		}
		var chData []string
		chData = append(chData, ch)
		chData = append(chData, content)
		results = append(results, chData)
		return nil
	}

	if err := readAll(p.Opf.Spine.Items); err != nil {
		return nil, err
	}

	return results, nil
}
