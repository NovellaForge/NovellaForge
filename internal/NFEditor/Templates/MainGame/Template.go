package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"go.novellaforge.dev/novellaforge/pkg/NFData"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction"
	"go.novellaforge.dev/novellaforge/pkg/NFFunction/DefaultFunctions"
	"go.novellaforge.dev/novellaforge/pkg/NFLayout/DefaultLayouts"
	"go.novellaforge.dev/novellaforge/pkg/NFLog"
	"go.novellaforge.dev/novellaforge/pkg/NFSave"
	"go.novellaforge.dev/novellaforge/pkg/NFScene"
	"go.novellaforge.dev/novellaforge/pkg/NFWidget/DefaultWidgets"
	"log"
	"os"
	"time"
	config "{{.LocalConfig}}"
	ExampleFunctions "{{.LocalFunctions}}"
	ExampleLayouts "{{.LocalLayouts}}"
	ExampleWidgets "{{.LocalWidgets}}"
)

func main() {
	//These functions allow specifying which functions, layouts, and widgets are available to the game
	DefaultFunctions.Import()
	DefaultLayouts.Import()
	DefaultWidgets.Import()
	ExampleFunctions.Import()
	ExampleLayouts.Import()
	ExampleWidgets.Import()
	//gameApp is the main app for the game to run on, when in a desktop environment this is the window manager that allows multiple windows to be open
	// The ID needs to be unique to the game, it is used to store preferences and other things if you overlap with another game, you may have issues with preferences and other things
	gameApp := app.NewWithID("com.novellaforge." + config.GameName)
	//window is the main window for the game, this is where the game is displayed and scenes are rendered
	window := gameApp.NewWindow(config.GameName + " " + config.Version)

	userHome, err := os.UserHomeDir()
	if err != nil {
		//NFInterface is a custom type that stores any data type mapping it to a string key for easy access there are two methods of declaring it,
		// the first is to use the NewNFInterface function like this before running set like you see below
		functionArgs := NFData.NewNFInterface()
		argErr := functionArgs.Set("Error", "Error Getting User Home Directory: "+err.Error())
		if argErr != nil {
			//The only time this error should exist is if you try to set or add a key that contains a period, as that is used to reference global or scene variables
			log.Println(argErr.Error())
		}

		/*
			//The second method is to declare the key value in the function call like this. Please note that the first method is safer as it will catch any errors thrown by the Set function,
			//but this method allows you to declare multiple key value pairs in the function call
			functionArgs = NFData.NewNFInterface(NFData.NewKeyVal("Error", "Error Getting User Home Directory: "+err.Error()))
			//An example of setting multiple values at once (You can use the same format in the NewNFInterface function SetMulti is just a method to do it after the initial declaration)
			functionArgs.SetMulti(NFData.NewKeyVal("Error", "Error Getting User Home Directory: "+err.Error()), NFData.NewKeyVal("Test", "Test"))
		*/

		//When calling a function it returns two values, the first is the data returned by the function, the second is an error if there is one, if there is not it will be nil
		returnArgs, funcErr := DefaultFunctions.CustomError(window, functionArgs)
		if funcErr != nil {
			log.Println(err.Error())
		}
		//to get a value from the arguments you first declare the variable and its type before passing it in to the Get Function as a ref
		var message string
		//This function also returns an error but as you can see we have declared the err as _ which in go means it is ignored, you can ignore any value by using _
		//and it will prevent the compiler from throwing an error if you don't use the value
		//You can see this used a lot later throughout the code when we don't care about the return value of a function
		//Also as you may have picked up on there are two ways of declaring variables in go, the first is to use := which is shorthand for declaring and assigning a variable to the type of the value
		//The second is to use var and then the variable name and type, this is used when you want to declare a variable but not assign it a value
		//When using _ to ignore a value, you never declare the value so if all return values are ignored you can just use = function() instead of := function()
		//if you are only ignoring one of multiple return values you need to declare the variable or use := when calling the function to assign it, _ variables are not affected by this
		_ = returnArgs.Get("TestGetMessage", &message)
	}

	NFSave.Directory = gameApp.Preferences().StringWithFallback("savesDir", userHome+"/MyGames/"+config.GameName+"/Saves")
	NFLog.Directory = gameApp.Preferences().StringWithFallback("logDir", userHome+"/MyGames/"+config.GameName+"/Logs")
	err = NFLog.SetUp(window, NFLog.Directory)
	if err != nil {
		log.Fatal(err)
	}
	splashScreen := gameApp.Preferences().BoolWithFallback("splashScreen", true)
	startupSettings := gameApp.Preferences().BoolWithFallback("startupSettings", true)
	if startupSettings {
		ShowStartupSettings(window, NFScene.GetAll(), splashScreen)
	} else {
		ShowGame(window, NFScene.GetAll(), "MainMenu", splashScreen)
	}
	window.ShowAndRun()
}

func createSplashScreen() fyne.Window {
	if drv, ok := fyne.CurrentApp().Driver().(desktop.Driver); ok {
		splash := drv.CreateSplashWindow()
		splash.SetContent(container.NewVBox(
			widget.NewLabelWithStyle(config.GameName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Version: "+config.Version, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle("Developed By: "+config.Author, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle("Powered By: NovellaForge and Fyne", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		))
		return splash
	}
	return nil
}

func ShowGame(window fyne.Window, allScenes map[string]NFScene.Scene, scene string, screen bool) {
	window.SetMaster()
	if screen {
		splash := createSplashScreen()
		window.Hide()
		splash.Show()
		time.Sleep(1 * time.Second)
		splash.Close()
		window.Show()
	}
	currentApp := fyne.CurrentApp()
	window.SetFullScreen(currentApp.Preferences().BoolWithFallback("fullscreen", false))
	window.SetContent(container.NewVBox())
	window.SetTitle(config.GameName + " " + config.Version)
	window.SetCloseIntercept(func() {
		dialog.ShowConfirm("Are you sure you want to quit?", "Are you sure you want to quit?", func(b bool) {
			if b {
				window.Close()
			}
		}, window)
	})
	window.SetFixedSize(false)
	window.Resize(fyne.NewSize(800, 600))
	window.CenterOnScreen()
	window.SetMainMenu(fyne.NewMainMenu(
		fyne.NewMenu("Game",
			fyne.NewMenuItem("Settings", func() {
				dialog.ShowCustomWithoutButtons("Settings", CreateSettings(false, window), window)
			}),
			fyne.NewMenuItem("New", func() {
				functionArgs := NFData.NewNFInterface()
				_, _ = DefaultFunctions.NewGame(window, functionArgs)
			}),
			fyne.NewMenuItem("Load", func() {
				functionArgs := NFData.NewNFInterface()
				_, _ = DefaultFunctions.LoadGame(window, functionArgs)
			}),
			fyne.NewMenuItem("Save", func() {
				err := NFSave.Active.Save()
				if err != nil {
					functionArgs := NFData.NewNFInterface()
					_ = functionArgs.Set("Error", "Error Saving Game: "+err.Error())
					_, _ = DefaultFunctions.CustomError(window, functionArgs)
					return
				}
			}),
			fyne.NewMenuItem("Save As", func() {
				functionArgs := NFData.NewNFInterface()
				_, _ = DefaultFunctions.SaveAs(window, functionArgs)
			}),
			fyne.NewMenuItem("Quit", func() {
				functionArgs := NFData.NewNFInterface()
				_, _ = DefaultFunctions.Quit(window, functionArgs)
			}),
		),
		fyne.NewMenu("View",
			fyne.NewMenuItem("Fullscreen", func() {
				window.SetFullScreen(!window.FullScreen())
				//Set the app preferences to the current fullscreen state for next time
				gameApp := fyne.CurrentApp()
				gameApp.Preferences().SetBool("fullscreen", window.FullScreen())
			}),
			fyne.NewMenuItem("Console", func() {
				NFLog.ShowDialog(window)
			}),
		),
	))

	if len(allScenes) == 0 {
		//There are actually two ways to call a function, the first is the way we have been doing it so far, the second is to parse it as it happens in the scene parser
		//This parsing method looks for the function by name in our loaded functions and then calls it with the arguments passed to it
		functionArgs := NFData.NewNFInterface()
		_ = functionArgs.Set("Error", "No Scenes Found")
		_, _ = NFFunction.ParseAndRun(window, "Error", functionArgs) // This Error function is just DefaultFunctions.CustomError, this is how scenes can store functions in their data files
	}

	//Check if the startup scene exists
	if _, ok := allScenes[scene]; !ok {
		//If it doesn't exist, return an error
		functionArgs := NFData.NewNFInterface()
		_ = functionArgs.Set("Error", "Startup Scene Not Found")
		_, _ = DefaultFunctions.CustomError(window, functionArgs)
	}

	startupScene := allScenes[scene]

	//SceneParser parses the scene and returns a fyne.CanvasObject that can be added to the window
	sceneObject, err := startupScene.Parse(window)
	if err != nil {
		functionArgs := NFData.NewNFInterface()
		_ = functionArgs.Set("Error", "Error Parsing Scene: "+err.Error())
		_, _ = DefaultFunctions.CustomError(window, functionArgs)
	}
	window.SetContent(sceneObject)
}

func ShowStartupSettings(window fyne.Window, allScenes map[string]NFScene.Scene, splashScreen bool) {
	settingsBox := CreateSettings(true, window)
	var creditsModal *widget.PopUp
	creditsCloseButton := widget.NewButton("Close", func() {
		creditsModal.Hide()
	})
	creditsModal = widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabelWithStyle("Credits", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle(config.Credits, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			creditsCloseButton,
		),
		window.Canvas(),
	)
	creditsButton := widget.NewButton("Credits", func() {
		creditsModal.Show()
	})
	startButton := widget.NewButton("Start Game", func() {
		if splashScreen {
			go ShowGame(window, allScenes, "MainMenu", true)
		} else {
			go ShowGame(window, allScenes, "MainMenu", false)
		}
	})
	settingsBox.(*fyne.Container).Add(container.NewVBox(creditsButton, startButton))
	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(300, 400))
	window.CenterOnScreen()
	window.SetContent(settingsBox)
	window.SetTitle("Startup Settings")
}

func CreateSettings(isStartup bool, window fyne.Window) fyne.CanvasObject {
	currentApp := fyne.CurrentApp()
	fullScreenBinding := binding.NewBool()
	err := fullScreenBinding.Set(currentApp.Preferences().BoolWithFallback("fullscreen", false))
	if err != nil {
		panic(err)
	}
	fullScreenBinding.AddListener(binding.NewDataListener(func() {
		//Set the app preferences to the current fullscreen state for next time
		fullscreen, err := fullScreenBinding.Get()
		if err != nil {
			panic(err)
		}
		currentApp.Preferences().SetBool("fullscreen", fullscreen)
		if !isStartup {
			window.SetFullScreen(fullscreen)
		}
	}))
	splashScreenBinding := binding.NewBool()
	err = splashScreenBinding.Set(currentApp.Preferences().BoolWithFallback("splashScreen", true))
	if err != nil {
		panic(err)
	}
	splashScreenBinding.AddListener(binding.NewDataListener(func() {
		//Set the app preferences to the current fullscreen state for next time
		splashScreen, err := splashScreenBinding.Get()
		if err != nil {
			panic(err)
		}
		currentApp.Preferences().SetBool("splashScreen", splashScreen)
	}))
	startUpSettingBinding := binding.NewBool()
	err = startUpSettingBinding.Set(currentApp.Preferences().BoolWithFallback("startupSettings", true))
	if err != nil {
		panic(err)
	}
	startUpSettingBinding.AddListener(binding.NewDataListener(func() {
		//Set the app preferences to the current fullscreen state for next time
		startUpSetting, err := startUpSettingBinding.Get()
		if err != nil {
			panic(err)
		}
		currentApp.Preferences().SetBool("startupSettings", startUpSetting)
	}))
	musicBinding := binding.NewFloat()
	err = musicBinding.Set(currentApp.Preferences().FloatWithFallback("musicVolume", 100))
	if err != nil {
		panic(err)
	}
	musicBinding.AddListener(binding.NewDataListener(func() {
		//Set the app preferences to the current fullscreen state for next time
		musicVolume, err := musicBinding.Get()
		if err != nil {
			panic(err)
		}
		currentApp.Preferences().SetFloat("musicVolume", musicVolume)
		//TODO Add the audio manager and set the music volume
	}))
	effectsBinding := binding.NewFloat()
	err = effectsBinding.Set(currentApp.Preferences().FloatWithFallback("effectsVolume", 100))
	if err != nil {
		panic(err)
	}
	effectsBinding.AddListener(binding.NewDataListener(func() {
		//Set the app preferences to the current fullscreen state for next time
		effectsVolume, err := effectsBinding.Get()
		if err != nil {
			panic(err)
		}
		currentApp.Preferences().SetFloat("effectsVolume", effectsVolume)
		//TODO Add the audio manager and set the effects volume
	}))
	SettingsScrollBox := container.NewScroll(
		widget.NewForm(
			widget.NewFormItem("Music Volume", widget.NewSlider(0, 100)),
			widget.NewFormItem("Effects Volume", widget.NewSlider(0, 100)),
			widget.NewFormItem("Fullscreen", widget.NewCheckWithData("", fullScreenBinding)),
			widget.NewFormItem("Show Settings on Startup", widget.NewCheckWithData("", startUpSettingBinding)),
			widget.NewFormItem("Show Splash Screen on Startup", widget.NewCheckWithData("", splashScreenBinding)),
		),
	)
	SettingsScrollBox.SetMinSize(fyne.NewSize(300, 50))
	settingsBox := container.NewVBox(
		widget.NewLabelWithStyle("Startup Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle(config.GameName, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		widget.NewLabelWithStyle("Version: "+config.Version, fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		widget.NewLabelWithStyle("Author: "+config.Author, fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		widget.NewLabelWithStyle("Game Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		SettingsScrollBox,
		layout.NewSpacer(),
	)
	return settingsBox
}
