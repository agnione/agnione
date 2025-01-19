//	 Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 26/01/2024
//
//	 Copyright     :   © 2024 D. Ajith Nilantha de Silva contact@agnione.net
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
//	 Class/module  :   Kandy Application Framework - Core PlugIn Implementation
//
//	 Objective     :   Define common package for Application that work as a container for the business objects.
//						This package will to export plugins/libraries to the business objects, so that will helps
//						business objects to implements it's features by using the framework support
//
// #######################################################################################
//	Author                        	Date        	Action      	Description
// ---------------------------------------------------------------------------------------------------------------
//
//	 	Ajith de Silva		26/01/2024	Created 	Created the initial version
//	 	Ajith de Silva		29/01/2024	Updated 	Defined functions with parameters & return values
//	 	Ajith de Silva		29/01/2024	Updated 	Updated the WSMonitor as library
//	 	Ajith de Silva		01/03/2024	Updated 	Added Get_AppUnit function to return the application unit instance
//	 	Ajith de Silva		05/03/2024	Updated 	Added Get_Mailer function to return the mailer instance
// #######################################################################################

package agni

import (
	aap "agnione/v1/src/aau/iappunit" /// import the unit interface
	"errors"
	"fmt"
	"strings"

	ihttp "agnione/v1/src/afplugins/http/iahttpclient" /// import the http interface

	iws "agnione/v1/src/afplugins/websocket/iawsclient" /// import the wsclient interface

	atypes "agnione/v1/src/appfm/types"

	zutls "agnione.appfm/src/utils"
)

func (app *AgniApp) Get_Plugin_Config(pPluginCategory string, pPlugins *[]atypes.PlugIn, pType *string) (*atypes.PlugIn,error) {
	
	for _, _plugin := range *pPlugins {
		if strings.ToLower(_plugin.Type) == *pType {
			if _plugin.Enable==1{
				return &_plugin,nil
			}else{
				return nil,errors.New("Plugin " + pPluginCategory + "(" + *pType + ") is not enable")
			}
		}
	}

	return nil,errors.New("Plugin " + pPluginCategory + "(" + *pType + ") is NOT found")
}



func (app *AgniApp) Get_WSClient(pType *string) (iws.IAWSClient, error) {

	if len(*pType)==0{
		*pType="default"
	}
	
	_plugin_config,_err:=app.Get_Plugin_Config("websocket", &app.coreconfig.Plugins.Websocket,pType)
	
	defer func ()  {
		_plugin_config=nil
	}()
	
	if _err!=nil{
		return nil,_err
	}
	
	if _plugin_config.Enable != 1 {
		return nil, errors.New("websocket plug-in is disabled")
	}

	_path:=*app.base_path + _plugin_config.Path + _plugin_config.Name
	_iPlugIn, _err := zutls.Get_WSClientPlugIn(&_plugin_config.Ifname, &_path)
	if _err != nil {
		return nil, _err
	}else{
		return _iPlugIn, nil
	}
}

func (app *AgniApp) Get_RESTClient(pType *string) (ihttp.IAHTTPClient, error) {

	if len(*pType)==0{
		*pType="default"
	}
	
	_plugin_config,_err:=app.Get_Plugin_Config("http", &app.coreconfig.Plugins.HTTP,pType)
	
	defer func ()  {
		_plugin_config=nil
	}()
	
	if _err!=nil{
		return nil,_err
	}
	
	if _plugin_config.Enable != 1 {
		return nil, errors.New("http plug-in is disabled")
	}
	
	_path:=*app.base_path + _plugin_config.Path +  _plugin_config.Name
	if _iPlugIn, _err := zutls.Get_HTTPClientPlugIn(&_plugin_config.Ifname, &_path); _err != nil {
		return nil, _err
	} else {
		return _iPlugIn, nil
	}
}

func (app *AgniApp) Get_AppUnit(pAppUnit *int) (aap.IAppUnit, error) {
	
	if app.appconfig.Appunits[*pAppUnit].Enable != 1 {
		return nil, fmt.Errorf("http plug-in is disabled")
	}

	_iPlugIn, _err := zutls.Get_AppUnit(&app.appconfig.Appunits[*pAppUnit].Path)
	if _err != nil {
		return nil, _err
	}

	return _iPlugIn, nil
}

func (app *AgniApp) Units_List() ([]atypes.Appunit, error){

	if app.appconfig==nil{
		return nil,errors.New("application configuration is not initialized")
	}
	
	return app.appconfig.Appunits,nil
}

func (app *AgniApp) Unit_Stop(pUnitName *string,pForce bool)(bool,error){

	return true,nil
}

func (app *AgniApp) Unit_Start(pUnitName *string)(bool,error){

	return true,nil
}

func (app *AgniApp) Unit_Restart(pUnitName *string,pForce bool)(bool,error){

	return true,nil
}

func (app *AgniApp) Unit_Status(pUnitName *string)(*atypes.AppUnitInfo,error){

	if len(app.appUnits)==0 {
		return nil,errors.New("no application unit loaded")
	}
	
	_uinfo:=new(atypes.AppUnitInfo)
	_bfound:=false
	
	for _,_unit:=range app.appUnits{
		if _unit.Status().Info.Name==*pUnitName{
			_bfound=true
			_uinfo=&atypes.AppUnitInfo{
				Info: _unit.Status().Info,
				Req_Handled: _uinfo.Req_Handled,
				Req_Failed: _uinfo.Req_Failed,
				Routines: _uinfo.Routines,
			}
			break
		}
	}
	if !_bfound{
		return nil,errors.New("no matching application unit loaded. (" + *pUnitName + ")")
	}else{
		return _uinfo,nil
	}
}