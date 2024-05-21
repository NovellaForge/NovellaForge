package main

import (
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFFunction/DefaultFunctions"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFLayout"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFLayout/DefaultLayouts"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFWidget"
	"go.novellaforge.dev/novellaforge/pkg/NFData/NFObjects/NFWidget/DefaultWidgets"
)

func init() {
	DefaultFunctions.Import()
	DefaultWidgets.Import()
	DefaultLayouts.Import()
	//Add some form of function from the asset pack you want to import here
	//i.e ExampleAssetPack.Import()
	//Otherwise go fmt and some ide's may remove the function.
	//Most asset packs should register their assets on init,
	//but if they do not you will need to run their registration here to make sure
	//They are added to the registered assets before the export is run for all registered.

}

func main() {
	NFLayout.ExportRegistered()
	NFWidget.ExportRegistered()
	NFFunction.ExportRegistered()
	/*
		!~~~~~~~~~~~IMPORTANT~~~~~~~~~~~~~!
		NO NEW EXPORT REGISTERS OR ANY OTHER CODE IS NEEDED HERE
		JUST ADD THE IMPORTS AND THEY WILL BE EXPORTED
		AS LONG AS THEY WERE REGISTERED CORRECTLY

		THIS FILE IS PURELY FOR EDITOR USE AND WILL NOT BE INCLUDED IN THE GAME WHEN IT IS BUILT
		RUNNING THIS MAIN FUNCTION EXPORTS THE NEEDED FILES FOR THE EDITOR TO ASSIST YOU WITH
		POPULATING THE DATA FOR WIDGETS LAYOUTS AND FUNCTIONS
		!~~~~~~~~~~~IMPORTANT~~~~~~~~~~~~~!
	*/
}
