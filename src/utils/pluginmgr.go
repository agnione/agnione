/*
*****************************************************************************************************
# Author        :   D. Ajith Nilantha de Silva  | 02/01/2024

# Copyright     :   © 2024 D. Ajith Nilantha de Silva contact@agnione.net
						Licensed under the Apache License, Version 2.0 (the "License");
						you may not use this file except in compliance with the License.
						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

						Unless required by applicable law or agreed to in writing, software
						distributed under the License is distributed on an "AS IS" BASIS,
						WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
						See the License for the specific language governing permissions and
						limitations under the License.

# Class/module  :   plugin manager

# Objective     :   Package functions for plugin management
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
	"errors"
	"fmt"
	"plugin"

	aau "agnione/v1/src/aau/iappunit"                  /// import the unit interface
	ihttp "agnione/v1/src/afplugins/http/iahttpclient" /// import the http client interface

	iws "agnione/v1/src/afplugins/websocket/iawsclient" /// import the wsclient interface
)

// load_plugin loads the plugin of given plugin name and given interface name
// Returns plugin.Symbol and nil if success. Unless nil and error
func load_plugin(interface_name string, plugin_name *string) (plugin.Symbol, error) {

	// can handle symbolic link, but will no follow the link

	/// load module
	/// 1. open the so file to load the symbols
	_plugin, _err := plugin.Open(*plugin_name)
	
	defer func(){
		_plugin=nil
	}()
	
	if _err != nil {
		return nil, _err
	}

	/// Load the plugin symbol matching to the given interface
	_symPlugIn, _err := _plugin.Lookup(interface_name)
	if _err != nil {
		return nil, _err
	}

	return _symPlugIn, nil
}


// Get_HTTPPlugIn loads the http client library plugin from given file name and interface name
// Returns IAHTTPClient and nil if success. Unless nil and error
func Get_HTTPClientPlugIn(interface_name *string, plugin_filename *string) (ihttp.IAHTTPClient, error) {

	_symClient, _err := load_plugin(*interface_name, plugin_filename)
	
	defer func(){
		_symClient=nil

	}()
	
	if _err != nil {
		return nil, _err
	}

	/// 2. look up a symbol (an exported function or variable)
	/// in this case, variable Greeter

	/// 3. Assert that loaded symbol is of a desired type
	/// in this case interface type Greeter (defined above)
	var _iHTTPClientLib ihttp.IAHTTPClient
	defer func(){
		_iHTTPClientLib=nil
	}()
	
	_iHTTPClientLib, ok := _symClient.(ihttp.IAHTTPClient)
	if !ok {
		return nil, fmt.Errorf("unexpected type from module symbols %v", _iHTTPClientLib)
	}

	_ihttpClient := _iHTTPClientLib.New() /// creates the instance
	if _ihttpClient == nil {
		return nil, errors.New("failed to creates a instance of " +  *interface_name)
	}

	_httpClient := _ihttpClient.(ihttp.IAHTTPClient)

	if _httpClient == nil {
		return nil, errors.New("failed to creates a instance from interface of " +  *interface_name)
	}

	return _httpClient, nil

}

// Get_WSClientTPlugIn loads the http client library plugin from given file name and interface name
// Returns IZWSClient and nil if success. Unless nil and error
func Get_WSClientPlugIn(interface_name *string, plugin_filename *string) (iws.IAWSClient, error) {

	_symClient, _err := load_plugin(*interface_name, plugin_filename)
	
	defer func(){
		_symClient=nil
	}()
	
	if _err != nil {
		return nil, _err
	}

	/// 2. look up a symbol (an exported function or variable)
	/// in this case, variable Greeter

	/// 3. Assert that loaded symbol is of a desired type
	/// in this case interface type Greeter (defined above)
	var _iWSClient iws.IAWSClient
	defer func(){
		_iWSClient=nil
	}()
	
	_iWSClient, ok := _symClient.(iws.IAWSClient)
	if !ok {
		return nil, fmt.Errorf("unexpected type from module symbols %v", _iWSClient)
	}

	_iwsc := _iWSClient.New() /// creates the instance
	if _iwsc == nil {
		return nil, errors.New("failed to creates a instance of " +  *interface_name)
	}

	_wsClient := _iwsc.(iws.IAWSClient)

	if _wsClient == nil {
		return nil, errors.New("failed to creates a instance from interface of " +  *interface_name)
	}

	return _wsClient, nil
}

// Get_WSClientTPlugIn loads the http client library plugin from given file name and interface name
// Returns IAHTTPClient and nil if success. Unless nil and error
func Get_AppUnit(plugin_filename *string) (aau.IAppUnit, error) {

	_symClient, _err := load_plugin("IAppUnit", plugin_filename)
	
	defer func(){
		_symClient=nil
	}()
	
	if _err != nil {
		return nil, _err
	}

	/// 2. look up a symbol (an exported function or variable)
	/// in this case, variable Greeter

	/// 3. Assert that loaded symbol is of a desired type
	/// in this case interface type Greeter (defined above)
	var _appUnitLib aau.IAppUnit
	
	defer func(){
		_appUnitLib=nil
	}()
	
	_appUnitLib, ok := _symClient.(aau.IAppUnit)
	if !ok {
		return nil, fmt.Errorf("unexpected type from module symbols %v", _appUnitLib)
	}

	_appUnit := _appUnitLib.New() /// creates the instance
	if _appUnit == nil {
		return nil, errors.New("failed to creates a instance of IAppUnit")
	}

	_iappUnit := _appUnit.(aau.IAppUnit)

	if _iappUnit == nil {
		return nil, errors.New("failed to creates a instance from interface of IAppUnit")
	}

	return _iappUnit, nil

}
