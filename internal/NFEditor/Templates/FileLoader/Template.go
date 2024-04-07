package main

import (
	"embed"
)

//---***DO NOT TOUCH THIS FILE UNLESS YOU KNOW FOR SURE WHAT YOU ARE DOING***---//
//---***THIS FILE IS MANAGED BY THE EDITOR AND ALTHOUGH IT SHOULD BE OVERWRITTEN IF THINGS WENT WRONG***---//
//---***IT IS NOT RECOMMENDED TO MANUALLY EDIT THIS FILE***---//

//go:embed {{.ProjectName}}.NFProject
var ConfigFile embed.FS

//These two options will be enabled if the user selects the corresponding options in the editor, please note that ANY
//References to data or assets should be done through these variables if you are using embedded assets or data
//Failure to do so will result in the editor not being able to find the files and errors will occur
//They are empty by default and will be populated by the editor if the user selects the corresponding options

var DataFolder embed.FS

var AssetsFolder embed.FS
