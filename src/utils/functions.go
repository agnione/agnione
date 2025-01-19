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

# Class/module  :   utulity functions

# Objective     :   package collection of utility functions for the framework

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
	"fmt"
	"math"
	"os/exec"
	"strings"
)

// FormatByteSize formats the bytes into matching size.
// Returns the formatted string
func FormatByteSize(byte_size uint64) string {
	bf := float64(byte_size)
	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(bf) < 1024.0 {
			return fmt.Sprintf("%3.1f%sB", bf, unit)
		}
		bf /= 1024.0
	}
	return fmt.Sprintf("%.1fYiB", bf)
}




// RunCMD is a simple wrapper around terminal commands
func run_cmd(path string, args []string) (out string, err error) {

    cmd := exec.Command(path, args...)

    var b []byte
	
	defer func(){
		if _r:=recover();_r!=nil{
			fmt.Println("Recovered panic ",_r)
			_r=nil
		}
		
		b=nil
		cmd=nil
	}()
	
    b, err = cmd.CombinedOutput()
    return string(b),err
}



// Execute_Command executes the given OS command and returns result
func Execute_Command(command string) (string, error) {
	if len(command)==0{
		return "",fmt.Errorf("command can not be empty")
	}
	
	_params:=strings.Split(command, " ")
	return run_cmd(_params[0], _params[1:])

	
	
}
