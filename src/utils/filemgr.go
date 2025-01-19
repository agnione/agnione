/*
*****************************************************************************************************
# Author        :   D. Ajith Nilantha de Silva  contact@agnione.net  | 02/01/2024

# Copyright     :   © 2024 D. Ajith Nilantha de Silva contact@agnione.net
						Licensed under the Apache License, Version 2.0 (the "License");
						you may not use this file except in compliance with the License.
						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

						Unless required by applicable law or agreed to in writing, software
						distributed under the License is distributed on an "AS IS" BASIS,
						WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
						See the License for the specific language governing permissions and
						limitations under the License.

# Class/module  :   file manager

# Objective     :   Define functions to handle file management.
#######################################################################################################
# Author                        Date        Action      Description
#------------------------------------------------------------------------------------------------------
# Ajith de Silva				02/01/2024	Created 	Created the initial version
# Ajith de Silva				03/01/2024	Updated 	Defined functions with parameters & return values
#######################################################################################################
******************************************************************************************************
*/

package utils

import (
	atypes "agnione/v1/src/appfm/types"
	"bufio"
	"os"
)

// IsFileExist checks if the given file exists.
// Returns true if file exists. Unless false
func IsFileExist(filename *string) bool {
	_, err := os.Stat(*filename)
	if err != nil {
		return false
	} else {
		return true
	}
}

// Get_FileInfo returns the information of the given file in format of FileInfo struct
func Get_FileInfo(pFileName *string) (*atypes.FileInfo, error) {

	/// try to read teh file information
	fileInfo, _err := os.Lstat(*pFileName)
	if _err != nil {
		return nil, _err
	}

	/// set the related properties
	_fileInfo := &atypes.FileInfo{}
	_fileInfo.Name = fileInfo.Name()
	_fileInfo.ModTime = fileInfo.ModTime()
	_fileInfo.Size = fileInfo.Size()
	_fileInfo.IsDir = fileInfo.IsDir()
	return _fileInfo, nil
}

// Get_File_Content returns the content of the given file.
// Returns []byte and nil if file exists. Unless returns nil and error
func Get_File_Content(pFileName *string) (*[]byte, error) {
	if _data, _err := os.ReadFile(*pFileName); _err != nil {
		return nil, _err
	} else {
		return &_data, nil
	}
}

// Get_File_Content returns the content of the given file.
// Returns []byte and nil if file exists. Unless returns nil and error
func GetFileContentLines(pFileName *string) (*[]string, error) {
	if _file, _err := os.OpenFile(*pFileName, os.O_RDONLY, os.ModePerm); _err != nil {
		return nil, _err
	} else {
		defer _file.Close()
		lines := make([]string, 0)
		_scanner := bufio.NewScanner(_file)
		for _scanner.Scan() {
			lines = append(lines, _scanner.Text())
		}
		return &lines, nil
	}
}

// WriteFileContent writes the given content []byte to the given filename.
// Returns true and nil if file exists. Unless returns nil and error
func WriteFileContent(pFileName *string, pData *[]byte) (bool, error) {
	if _err := os.WriteFile(*pFileName, *pData, 0644); _err != nil {
		return false, _err
	} else {
		return true, nil
	}
}

// GetFilePtr returns the file pointer of the given file
// Returns valid file pointer and nil if file exists.
// Unless returns nil and error
func GetFilePtr(pFileName *string) (*os.File, error) {
	_file, _err := os.Open(*pFileName)
	if _err != nil {
		return nil, _err
	} else {
		return _file, nil
	}
}
