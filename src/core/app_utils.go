//
//#################################################################################################################
// Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 26/01/2024
//
// Copyright     :   © 2024 D. Ajith Nilantha de Silva contact@agnione.net
//						Licensed under the Apache License, Version 2.0 (the "License");
//						you may not use this file except in compliance with the License.
//						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
//						Unless required by applicable law or agreed to in writing, software
//						distributed under the License is distributed on an "AS IS" BASIS,
//						WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//						See the License for the specific language governing permissions and
//						limitations under the License.
//
// Class/module  :   Kandy Application Framework - utility functions
//
// Objective     :   Define common utility functions for Application that work.
//					This package will to export app units objects, so that will helps
//					business objects to implements it's features by using the framework support
//################################################################################################################
// Author                        Date        Action      Description
//------------------------------------------------------------------------------------------------------
// Ajith de Silva				26/01/2024	Created 	Created the initial version
// Ajith de Silva				29/01/2024	Updated 	Defined functions with parameters & return values
// Ajith de Silva				29/01/2024	Updated 	Updated the WSMonitor as library
// Ajith de Silva				03/04/2024	Added	 	Execute_Command method to execute the OS command
//#################################################################################################################
//

package agni

import (
	apptypes "agnione/v1/src/appfm/types"
	"os"

	autls "agnione.appfm/src/utils"
)

// Get_FileInfo returns the information of the given file in format of FileInfo struct
func (app *AgniApp) Get_FileInfo(pFileName *string) (*apptypes.FileInfo, error) {
	return autls.Get_FileInfo(pFileName)
}

// Get_File_Content returns the file content of the given file name.
//
// Returns file content []bytes,nil if successful.
//
// Unless returns nil and error
func (app *AgniApp) Get_File_Content(pFileName *string) (*[]byte, error) {
	return autls.Get_File_Content(pFileName)
}

// GetFilePtr returns the file pointer of the given file.
//
// Returns valid file pointer and nil if successful.
//
// Unless returns nil and error
func (app *AgniApp) GetFilePtr(pFileName *string) (*os.File, error) {
	return autls.GetFilePtr(pFileName)
}

// GetFileContentLines returns the file []content string of the given file name.
//
// Returns file content *[]string,nil if successful.
//
// Unless returns nil and error
func (app *AgniApp) Get_FileContent_Lines(pFileName *string) (*[]string, error) {
	return autls.GetFileContentLines(pFileName)
}

// WriteFileContent writes the given content []byte to the given filename.
// Returns true and nil if file exists. Unless returns nil and error
func (app *AgniApp) Write_FileContent(pFileName *string, pData *[]byte) (bool, error) {
	return autls.WriteFileContent(pFileName, pData)
}

// Execute_Command executes the given OS command and returns result
//
// command string parameter - the command to execute with arguments
// Returns result of the command execution as string.
// If failed then return the error
func (app *AgniApp) Execute_Command(pCommand *string) (string, error) {
	return autls.Execute_Command(*pCommand)
}

// / loadMainConfiguration loads the main configuration of the framework
func (app *AgniApp) LoadCoreConfiguration(pFilename *string) (*apptypes.FMConfig, error) {
	return autls.LoadCoreConfiguration(pFilename)
}

/*
// / loadMainConfiguration loads the main configuration of the framework
func (app *AgniApp) LoadMainConfiguration(filename *string) (*apptypes.MainConfig, error) {
	return autls.LoadMainConfiguration(filename)
}
	*/


// / loadAppConfiguration loads the application specific configuration
func (app *AgniApp) LoadAppConfiguration(pFilename *string) (*apptypes.AppConfig, error) {
	return autls.LoadAppConfiguration(pFilename)
}

