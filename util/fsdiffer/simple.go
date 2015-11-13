// Copyright 2015 Simone Gotti
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package fsdiffer

import (
	"os"
	"path/filepath"
)

type SimpleFSDiffer struct {
	sourceDir string
	destDir   string
}

type fileInfo struct {
	Path string
	os.FileInfo
}

func NewSimpleFSDiffer(sourceDir string, destDir string) *SimpleFSDiffer {
	return &SimpleFSDiffer{sourceDir: sourceDir, destDir: destDir}
}

// Creates the FSChanges between sourceDir and destDir.
// To detect if a file was changed it checks the file's size and mtime (like
// rsync does by default if no --checksum options is used)
func (s *SimpleFSDiffer) Diff() (FSChanges, error) {
	changes := FSChanges{}
	sourceFileInfos := make(map[string]fileInfo)
	destFileInfos := make(map[string]fileInfo)
	err := filepath.Walk(s.sourceDir, fsWalker(sourceFileInfos))
	if err != nil {
		return nil, err
	}
	err = filepath.Walk(s.destDir, fsWalker(destFileInfos))
	if err != nil {
		return nil, err
	}

	for _, destInfo := range destFileInfos {
		relpath, _ := filepath.Rel(s.destDir, destInfo.Path)
		sourceInfo, ok := sourceFileInfos[filepath.Join(s.sourceDir, relpath)]
		if !ok {
			changes = append(changes, &FSChange{Path: relpath, ChangeType: Added})
		} else {
			if sourceInfo.Size() != destInfo.Size() || sourceInfo.ModTime().Before(destInfo.ModTime()) {
				changes = append(changes, &FSChange{Path: relpath, ChangeType: Modified})
			}
		}
	}
	for _, infoA := range sourceFileInfos {
		relpath, _ := filepath.Rel(s.sourceDir, infoA.Path)
		_, ok := destFileInfos[filepath.Join(s.destDir, relpath)]
		if !ok {
			changes = append(changes, &FSChange{Path: relpath, ChangeType: Deleted})
		}
	}
	return changes, nil
}

func fsWalker(fileInfos map[string]fileInfo) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		fileInfos[path] = fileInfo{Path: path, FileInfo: info}
		return nil
	}
}
