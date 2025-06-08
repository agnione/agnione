// utils package provides all the utility functions for AgniOne application framework
//
// Configuration functions:
//			- LoadMainConfiguration
//			- LoadAppConfiguration
//			- Stop
// File management functions:
//			- IsFileExist
//			- Get_FileInfo
//			- Get_File_Content
//			- GetFileContrntLines
//			- WriteFileContent
//			- GetFilePtr
// Common functions:
//			- FormatByteSize
//			- Execute_Command
// Plugin functions:
//			- load_plugin
//			- Get_MailerPlugIn
//			- Get_MQClientPlugIn
//			- Get_MQClusterClientPlugIn
//			- Get_HTTPClientPlugIn
//			- Get_WSClientPlugIn
//			- Get_LoggerPlugIn
/*
########################################################################################

Author        :   D. Ajith Nilantha de Silva  | 02/01/2024

	Copyright     :   © 2024 D. Ajith Nilantha de Silva contact@agnione.net
						Licensed under the Apache License, Version 2.0 (the "License");
						you may not use this file except in compliance with the License.
						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

						Unless required by applicable law or agreed to in writing, software
						distributed under the License is distributed on an "AS IS" BASIS,
						WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
						See the License for the specific language governing permissions and
						limitations under the License.

	Class/module  :   config manager

	Objective     :   Define functions to handle configuration loading from the given file name.

	Configuration file should be in TOML format

#########################################################################################
	Author                 	Date        	Action      	Description
-----------------------------------------------------------------------------------------------------------------
	Ajith de Silva		02/01/2024	Created 	Created the initial version
	Ajith de Silva		03/01/2024	Updated 	Defined functions with parameters & return values
########################################################################################
*/
package utils

import (
	apptypes "agnione/v2/src/appfm/types"
	"encoding/json"
	"errors"
)

// / LoadCoreConfiguration loads the core configuration
func LoadCoreConfiguration(filename *string) (*apptypes.FMConfig, error) {

	_file, _err := GetFilePtr(filename)

	if _err != nil {
		return nil, _err
	}
	_config := &apptypes.FMConfig{}
	defer func() {
		_file.Close()
		_file = nil
		_err = nil
	}()

	_err = json.NewDecoder(_file).Decode(_config)

	if _err != nil {
		return nil, errors.New("Error decoding JSON data: " + _err.Error())
	}

	return _config, nil /// all good.
}

// / loadAppConfiguration loads the application configuration
func LoadAppConfiguration(filename *string) (*apptypes.AppConfig, error) {

	_file, _err := GetFilePtr(filename)

	if _err != nil {
		return nil, _err
	}

	_appConfig := &apptypes.AppConfig{}

	defer func() {
		_file.Close()
		_file = nil
		_err = nil
	}()

	_err = json.NewDecoder(_file).Decode(_appConfig)

	if _err != nil {
		return nil, errors.New("Error decoding JSON data: " + _err.Error())
	}
	return _appConfig, nil /// all good.
}
