# NovellaForge
An Open Source Visual Novel Creator In GO, UI built with [Fyne](https://fyne.io/)

## Mission Statement
NovellaForge aims to democratize the creation of visual novels by providing an open source game engine that requires zero programming knowledge for building a functional visual novel or other 2D games. This is accomplished by building a framework and editor to be fully extensible and modular, allowing full control over the game's source code. For advanced developers with an understanding of GO this flexibility enables the creation of games that go beyond the traditional visual novel genre.

## TODO 0.0.1: (These are not final and will be updated as needed)
- [x] Project creation and detection for open recent and continue last project
- [x] Basic game templates allowing for instant game creation
- [x] Dynamic type interface to allow loading scenes from file
- [x] Global, per scene, and per object properties to allow for dynamic content in a non-linear fashion
- [x] Simple logging extension and built in terminal for debugging
- 
- [ ] Easily Extensible save system to allow for saving and loading game data
- [ ] Scene editor
    - [ ] Add in the ability to add and remove scenes
    - [x] Add in the ability to group scenes
    - [ ] ~~Add in the ability to change the order of scenes~~ (I sorted them alphabetically for now)
    - [ ] Add in the ability to add and remove widgets and containers and change all their relevant properties
    - [ ] Add in the ability to change the name of scenes
    - [ ] Possibly add in the ability to group widgets and layouts
    - [ ] Possibly Add in the ability to change the order of widgets and layouts
- [ ] Refactor Saving to use the new interface system and integrate it with default game templates
- [ ] Add in the build manager
    - [ ] Include embedding toggle to embed assets in the binary
    - [ ] Property manager needs to be able to fully edit all widget and container properties

## TODO 0.1.0:
- [ ] Add in a way to run the game from the editor for testing
- [ ] Scene Editor Preview Mode
    - [ ] This should add clickable elements to all elements that select them for editing in the properties
    - [ ] This should override buttons and other interactive elements to not be clickable

## TODO 1.0.0:
- [ ] Fully Documented in code and on wiki
- [ ] Add in the debug run mode
    - [ ] This should run the game with a debug flag enabled, that enables editing in certain widget that support it
- [ ] Add in the ability to open the project in the default IDE
- [ ] Add in the ability to open the project in the default file manager
- [ ] Add in the ability to open the project in the default terminal
- [ ] Add in the ability to change what IDE is used to open the project
- [ ] Add in the ability to generate a keystore for android builds

## TODO Future:
- [ ] Add in drag and drop widget support
- [ ] Add in IOS and Mac support for game builds and Mac support for editor functionality

## Feature Breakdown
- **User-Friendly GUI**: An intuitive interface that makes game creation as simple as possible.
- **Modular and Extensible Framework**: A flexible architecture that allows for custom modules and extensions.
- **Full Source Code Control**: Users have complete access to their game's source code, allowing for limitless customization.
- **Zero Programming Knowledge Required**: The basic functionality of the game engine can be used without any programming knowledge.
- **Performance and Reliability**: The game engine is designed to run smoothly and reliably, even on lower-end hardware.

## Installation Instructions
*Coming Soon*

## Custom Expansion Instructions
*Coming Soon*

## Contributing
We welcome contributions from the community. Whether it's reporting a bug, proposing a new feature, or writing code, every contribution helps us improve NovellaForge. Please see our contributing guide for more information.

## License
NovellaForge is open source and licensed under the [GNU General Public Licence V3](LICENSE).