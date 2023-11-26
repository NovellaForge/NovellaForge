package NFProject

import (
	"github.com/NovellaForge/NovellaForge/pkg/NFLayout"
	"github.com/NovellaForge/NovellaForge/pkg/NFScene"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
)

// MainGameTemplate is the template for the <project-name>.go file in the project directory it should contain a fyne app and window and the main menu set.
// Make sure you add the config import to and the package main line to the top of the file
const MainGameTemplate = `
import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFFunction"
	"github.com/NovellaForge/NovellaForge/pkg/NFLog"
	"github.com/NovellaForge/NovellaForge/pkg/NFSave"
	"github.com/NovellaForge/NovellaForge/pkg/NFScene"
	"log"
	"os"
	"time"
)

// TODO Move the splash and startup settings to a config file in the os.userConfigDir

func main() {
	//gameApp is the main app for the game to run on, when in a desktop environment this is the window manager that allows multiple windows to be open
	// The ID needs to be unique to the game, it is used to store preferences and other things if you overlap with another game, you may have issues with preferences and other things
	gameApp := app.NewWithID("com.novellaforge." + GameName)
	//window is the main window for the game, this is where the game is displayed and scenes are rendered
	window := gameApp.NewWindow(GameName + " " + GameVersion)

	userHome, err := os.UserHomeDir()
	if err != nil {
		_, _, _ = NFFunction.CustomError(window, map[string]interface{}{"message": "Error Getting User Home Directory: " + err.Error()})
	}

	NFSave.Directory = gameApp.Preferences().StringWithFallback("savesDir", userHome+"/MyGames/"+GameName+"/Saves")
	NFLog.Directory = gameApp.Preferences().StringWithFallback("logDir", userHome+"/MyGames/"+GameName+"/Logs")
	err = NFLog.SetUp()
	if err != nil {
		log.Fatal(err)
	}
	splashScreen := gameApp.Preferences().BoolWithFallback("splashScreen", true)
	startupSettings := gameApp.Preferences().BoolWithFallback("startupSettings", true)

	if splashScreen {
		go SplashScreenLoop(window)
	} else {
		window.Show()
	}

	if startupSettings {
		go ShowStartupSettings(window, NFScene.All)
	} else {
		//The main game window runs in a separate go routine so that waits using time.Sleep() do not block the UI
		//Keep in mind that creating additional go routines will require additional handling for pausing the game and other things
		go ShowGame(window, NFScene.All, StartupScene)
	}

	gameApp.Run()
}

func ShowGame(window fyne.Window, allScenes map[string]NFScene.Scene, scene string) {
	currentApp := fyne.CurrentApp()
	window.SetFullScreen(currentApp.Preferences().BoolWithFallback("fullscreen", false))
	window.SetContent(container.NewVBox())
	window.SetTitle(GameName + " " + GameVersion)
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
				_, _, _ = NFFunction.NewGame(window, nil)
			}),
			fyne.NewMenuItem("Load", func() {
				_, _, _ = NFFunction.LoadGame(window, nil)
			}),
			fyne.NewMenuItem("Save", func() {
				err := NFSave.Active.Save()
				if err != nil {
					_, _, _ = NFFunction.CustomError(window, map[string]interface{}{"message": "Error Saving Game: " + err.Error()})
					return
				}
			}),
			fyne.NewMenuItem("Save As", func() {
				_, _, _ = NFFunction.SaveAs(window, nil)
			}),
			fyne.NewMenuItem("Quit", func() {
				_, _, _ = NFFunction.Quit(window, nil)
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
		_, err, _ := NFFunction.Parse(window, "Error", map[string]interface{}{"message": "No Scenes Found"})
		if err != nil {
			panic(err)
		}
	}

	//Check if the startup scene exists
	if _, ok := allScenes[scene]; !ok {
		//The error dialog does not actually return anything, so we can ignore the data returned by the function
		_, _, _ = NFFunction.CustomError(window, map[string]interface{}{"message": "Startup Scene Not Found"})
	}

	startupScene := allScenes[scene]

	//SceneParser parses the scene and returns a fyne.CanvasObject that can be added to the window
	sceneObject, err := startupScene.Parse(window)
	if err != nil {
		_, _, _ = NFFunction.CustomError(window, map[string]interface{}{"message": "Error Parsing Scene: " + err.Error()})
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
			widget.NewLabelWithStyle(GameCredits, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			creditsCloseButton,
		),
		window.Canvas(),
	)
	creditsButton := widget.NewButton("Credits", func() {
		creditsModal.Show()
	})
	startButton := widget.NewButton("Start Game", func() {
		//The main game window runs in a separate go routine so that waits using time.Sleep() do not block the UI
		go ShowGame(window, allScenes, StartupScene)
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
		widget.NewLabelWithStyle(GameName, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		widget.NewLabelWithStyle("Version: "+GameVersion, fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		widget.NewLabelWithStyle("Author: "+GameAuthor, fyne.TextAlignLeading, fyne.TextStyle{Italic: true}),
		widget.NewLabelWithStyle("Game Settings", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		SettingsScrollBox,
		layout.NewSpacer(),
	)
	return settingsBox
}

func SplashScreenLoop(window fyne.Window) {
	drv := fyne.CurrentApp().Driver()
	for {
		//Check if the driver is ready
		if drv != nil {
			//If it is, break out of the loop
			break
		}
	}

	if drv, ok := drv.(desktop.Driver); ok {
		splash := drv.CreateSplashWindow()
		splashBox := container.NewVBox(
			widget.NewLabelWithStyle(GameName, fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
			widget.NewLabelWithStyle("Version: "+GameVersion, fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle("Created By: "+GameAuthor+" Using Novella Forge", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
			widget.NewLabelWithStyle("Powered By: Fyne", fyne.TextAlignCenter, fyne.TextStyle{Italic: true}),
		)
		splash.SetContent(splashBox)
		splash.Show()
		go func() {
			time.Sleep(time.Second * 3)
			splash.Close()
			window.Show()
		}()
	} else {
		//If splash screens are not supported, just show the window as soon as the driver is ready
		window.Show()
	}
}
`
const CustomFunctionTemplate = `
import . "github.com/NovellaForge/NovellaForge/pkg/NFFunction"
func init() {
	Register("ExampleFunction", ExampleFunction)
}
`
const ExampleFunctionTemplate = `
package handlers

import (
	"errors"
	"fyne.io/fyne/v2"
)

var ExampleGlobalVariable string  // this is a global variable, it will be accessible to all functions by importing the package
var examplePrivateVariable string // this is a private variable, it will only be accessible to this file unless it is returned

func ExampleFunction(window fyne.Window, args map[string]interface{}) (map[string]interface{}, map[string]fyne.CanvasObject, error) {
	//Check if it has a text argument
	if text, ok := args["text"]; ok {
		//Set both the global and private variables to the text
		ExampleGlobalVariable = text.(string) + " (global)"
		examplePrivateVariable = text.(string)
		//Return the text as a string
		return map[string]interface{}{"text": examplePrivateVariable}, nil, nil
	}
	//If it doesn't, return an error
	return nil, nil, errors.New("no text argument")

}
`
const CustomLayoutTemplate = `
package layouts
import . "DefaultGame/internal/layouts/handlers"
import . "github.com/NovellaForge/NovellaForge/pkg/NFLayout"
func init() {
	Register("ExampleMenu", ExampleLayout)
}
`
const ExampleLayoutTemplate = `
package handlers

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
)

func ExampleLayout(window fyne.Window, _ map[string]interface{}, children []*NFWidget.Widget) (fyne.CanvasObject, error) {
	vbox := container.NewVBox()
	for _, child := range children {
		widget, err := child.Parse(window)
		if err != nil {
			return nil, err
		}
		vbox.Add(widget)
	}
	return vbox, nil
}
`
const CustomWidgetTemplate = `
package widgets
import 	. "DefaultGame/internal/widgets/handlers"
import 	. "github.com/NovellaForge/NovellaForge/pkg/NFWidget"
func init() {
	Register("ExampleWidget", ExampleWidget)
}
`
const ExampleWidgetTemplate = `
import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/NovellaForge/NovellaForge/pkg/NFFunction"
	"github.com/NovellaForge/NovellaForge/pkg/NFWidget"
	"log"
)

func ExampleWidget(window fyne.Window, args map[string]interface{}, _ []*NFWidget.Widget) (fyne.CanvasObject, error) {
	if text, ok := args["text"]; ok {
		if action := args["action"]; ok {
			button := widget.NewButton(text.(string), func() {

				_, _, err := NFFunction.Parse(window, action.(string), args)
				if err != nil {
					return
				}
			})
			return button, nil
		}
		button := widget.NewButton(text.(string), func() {
			log.Println("Example Button Pressed: Global variable is: " + handlers.ExampleGlobalVariable)
		})
		return button, nil
	}
	return nil, errors.New("no text argument")
}
`

var MainMenuSceneTemplate = NFScene.Scene{
	Name: "MainMenu",
	Layout: NFLayout.Layout{
		Type: "VBox",
		Children: []NFWidget.Widget{
			{
				Type: "Label",
				Properties: map[string]interface{}{
					"Text": "Main Menu",
				},
			},
			{
				Type: "Button",
				Properties: map[string]interface{}{
					"Text":   "New Game",
					"Action": "NewGame",
				},
			},
			{
				Type: "Button",
				Properties: map[string]interface{}{
					"Text":   "Load Game",
					"Action": "LoadGame",
				},
			},
			{
				Type: "Button",
				Properties: map[string]interface{}{
					"Text":   "Settings",
					"Action": "Settings",
				},
			},
			{
				Type: "Button",
				Properties: map[string]interface{}{
					"Text":   "Quit",
					"Action": "Quit",
				},
			},
		},
	},
	Properties: nil,
}

var NewGameSceneTemplate = NFScene.Scene{
	Name: "NewGame",
	Layout: NFLayout.Layout{
		Type: "VBox",
	},
}
