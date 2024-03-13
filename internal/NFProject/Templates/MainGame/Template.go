package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFFunction"
	"github.com/NovellaForge/NovellaForge/pkg/NFFunction/DefaultFunctions"
	"github.com/NovellaForge/NovellaForge/pkg/NFLayout"
	"github.com/NovellaForge/NovellaForge/pkg/NFLayout/DefaultLayouts"
	"github.com/NovellaForge/NovellaForge/pkg/NFLog"
	"github.com/NovellaForge/NovellaForge/pkg/NFSave"
	"github.com/NovellaForge/NovellaForge/pkg/NFScene"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget/DefaultWidgets"
	"log"
	"os"
	config "{{.LocalConfig}}"
)

// TODO Fix the template to properly import the local config package when the game is built

func main() {
	DefaultFunctions.Import()
	DefaultLayouts.Import()
	DefaultWidgets.Import()
	//gameApp is the main app for the game to run on, when in a desktop environment this is the window manager that allows multiple windows to be open
	// The ID needs to be unique to the game, it is used to store preferences and other things if you overlap with another game, you may have issues with preferences and other things
	gameApp := app.NewWithID("com.novellaforge." + config.GameName)
	//window is the main window for the game, this is where the game is displayed and scenes are rendered
	window := gameApp.NewWindow(config.GameName + " " + config.GameVersion)

	userHome, err := os.UserHomeDir()
	if err != nil {
		_, _, _ = DefaultFunctions.CustomError(window, map[string]interface{}{"message": "Error Getting User Home Directory: " + err.Error()})
	}

	NFSave.Directory = gameApp.Preferences().StringWithFallback("savesDir", userHome+"/MyGames/"+config.GameName+"/Saves")
	NFLog.Directory = gameApp.Preferences().StringWithFallback("logDir", userHome+"/MyGames/"+config.GameName+"/Logs")
	err = NFLog.SetUp(window, NFLog.Directory)
	if err != nil {
		log.Fatal(err)
	}
	//splashScreen := gameApp.Preferences().BoolWithFallback("splashScreen", true)
	//startupSettings := gameApp.Preferences().BoolWithFallback("startupSettings", true)

	//TODO Setup settings that lead into splash screen and then game, with both splash screen and settings being optional

	TempGameScene := NFScene.Scene{
		Name: "TempGameScene",
		Layout: NFLayout.Layout{
			Type: "ExampleLayout",
			Children: []NFWidget.Widget{
				{
					Type: "ExampleWidget",
					Properties: map[string]interface{}{
						"message": "Hello World",
						"action":  "ExampleFunction",
					},
				},
			},
			Properties: nil,
		},
		Properties: nil,
	}
	scene, err := TempGameScene.Parse(window)
	if err != nil {
		_, _, _ = DefaultFunctions.CustomError(window, map[string]interface{}{"message": "Error Parsing Scene: " + err.Error()})
	}
	window.SetContent(scene)
	gameApp.Run()
}

func ShowGame(window fyne.Window, allScenes map[string]NFScene.Scene, scene string) {
	currentApp := fyne.CurrentApp()
	window.SetFullScreen(currentApp.Preferences().BoolWithFallback("fullscreen", false))
	window.SetContent(container.NewVBox())
	window.SetTitle(config.GameName + " " + config.GameVersion)
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
				_, _, _ = DefaultFunctions.NewGame(window, nil)
			}),
			fyne.NewMenuItem("Load", func() {
				_, _, _ = DefaultFunctions.LoadGame(window, nil)
			}),
			fyne.NewMenuItem("Save", func() {
				err := NFSave.Active.Save()
				if err != nil {
					_, _, _ = DefaultFunctions.CustomError(window, map[string]interface{}{"message": "Error Saving Game: " + err.Error()})
					return
				}
			}),
			fyne.NewMenuItem("Save As", func() {
				_, _, _ = DefaultFunctions.SaveAs(window, nil)
			}),
			fyne.NewMenuItem("Quit", func() {
				_, _, _ = DefaultFunctions.Quit(window, nil)
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

	if len(allScenes) < 1 {
		//All functions with the function parser can return a map[string]interface{} and an error, the map[string]interface{} is used to return data to the main function, in this case we do not need to use any data from the function so we can just ignore it
		// An _ is used to ignore the data returned by the function, other functions will need error handling as shown here,
		// This function is the same as the functions.NewError function, but it uses the parser instead of the function directly,
		// The error handling is not needed for this function as it will never return an error, this is just to show a very basic example of how to handle errors
		// All functions that use the parser will need to be added to either the default or custom functions map, and they must have all three return values
		// If you are importing or creating a custom package for functions, you should add an init function to the package that adds the functions to the map and takes in the map as a parameter, so you don't have to have a recursive import
		_, err, _ := NFFunction.ParseAndRun(window, "Error", map[string]interface{}{"message": "No Scenes Found"})
		if err != nil {
			panic(err)
		}
	}

	//Check if the startup scene exists
	if _, ok := allScenes[scene]; !ok {
		//The error dialog does not actually return anything, so we can ignore the data returned by the function
		_, _, _ = DefaultFunctions.CustomError(window, map[string]interface{}{"message": "Startup Scene Not Found"})
	}

	startupScene := allScenes[scene]

	//SceneParser parses the scene and returns a fyne.CanvasObject that can be added to the window
	sceneObject, err := startupScene.Parse(window)
	if err != nil {
		_, _, _ = DefaultFunctions.CustomError(window, map[string]interface{}{"message": "Error Parsing Scene: " + err.Error()})
	}
	window.SetContent(sceneObject)
}

func ShowStartupSettings(window fyne.Window, allScenes map[string]NFScene.Scene) {
	settingsBox := CreateSettings(true, window)
	var creditsModal *widget.PopUp
	creditsCloseButton := widget.NewButton("Close", func() {
		creditsModal.Hide()
	})
	creditsModal = widget.NewModalPopUp(
		container.NewVBox(
			widget.NewLabelWithStyle("Credits", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle(config.GameCredits, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			creditsCloseButton,
		),
		window.Canvas(),
	)
	creditsButton := widget.NewButton("Credits", func() {
		creditsModal.Show()
	})
	startButton := widget.NewButton("Start Game", func() {
		//The main game window runs in a separate go routine so that waits using time.Sleep() do not block the UI
		//go ShowGame(window, allScenes, StartupScene)
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
		widget.NewLabelWithStyle("Version: "+config.GameVersion, fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		widget.NewLabelWithStyle("Author: "+config.GameAuthor, fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		widget.NewLabelWithStyle("Game Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		SettingsScrollBox,
		layout.NewSpacer(),
	)
	return settingsBox
}
