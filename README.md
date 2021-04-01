<!-- PROJECT SHIELDS -->
<!--
*** I'm using markdown "reference style" links for readability.
*** Reference links are enclosed in brackets [ ] instead of parentheses ( ).
*** See the bottom of this document for the declaration of the reference variables
*** for contributors-url, forks-url, etc. This is an optional, concise syntax you may use.
*** https://www.markdownguide.org/basic-syntax/#reference-style-links
-->
[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![MIT License][license-shield]][license-url]
[![LinkedIn][linkedin-shield]][linkedin-url]


<!-- TABLE OF CONTENTS -->
<details open="open">
  <summary>Table of Contents</summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#configuration">Configuration</a></li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#acknowledgements">Acknowledgements</a></li>
  </ol>
</details>


<!-- ABOUT THE PROJECT -->
## About The Project

A client-server serial bus for Arduino keyboard to run scripts and activate LEDs according to them.
There are a lot of Arduino keyboards instructions and manuals out there. I focused on 2 things:
1. The option of running script and not just keystrokes or macros
2. Feedback that also turning on and off the LEDs according to whatever you config


### Built With

This section should list any major frameworks that you built your project using. Leave any add-ons/plugins for the acknowledgements section. Here are a few examples.
* [Go](https://golang.org)
* [Arduino](https://www.arduino.cc/)


<!-- GETTING STARTED -->
## Getting Started

### Prerequisites

* Of course you need to prepare your Arduino keyboard yourself. There are [plenty of tutorials](https://roboindia.com/tutorials/arduino-nano-digital-input-push-button/) on the internet of how to connect LEDs and buttons to it. make sure you mark the port of each LED and button.  
Note that currently the code support buttons connected to analog pins.
* You must have to install [Arduino IDE](https://www.arduino.cc/en/software) or similar in order to burn the .ino file to the Arduino. 
* Go 16 is required to build the project, earlier versions should work too.


### Installation

Arduino
1. Make sure you does these changes in the .ino file before uploading:
    1. change *analog_pins* array and *num_buttons* 
    2. change *btn_leds* array and *num_leds* to your corresponding used pins
    3. change the size of *block* array to the number of buttons
    4. change the size of *timers* array to the number of leds
2. compile and upload

Go script
```sh
go mod vendor
#if you want to pre-compile:
go build .
```

<!-- CONFIGURATIONS EXAMPLES -->
## Configurations

conf.json fields:
```conf.json
{
    "ledBoard": {
        "port": "<your PC port usually COM* in windows and /dev/ttyUSB* or /dev/ttyACM*>",
        "baud": 115200 //this is the same as in .ino no need to change.
    },
    "buttons": {
        "<.ino button array index +1>": {"cmd": "<command>"},
        //this will be execute using exec.Command.
        //On windows most of the commands should work, you might want to add "cmd \c" im the beginning
        //    but I recommend using .bat files.
        // linux will add "/bin/sh -c" for you.
        //example:
        "2": {"cmd": "cscript //nologo C:\\Users\\mrsag\\git\\ledboard\\scripts\\toggle_zoom_audio.js"},
    },
    "leds": {
      "<.ino led array index +1>": {
        "type" : "<activate rule type: toggle/cmd[/none]>",
        //if selected type: toggle
        "toggle": {
          "button": <button_number>
          "init": <true/false>
        },
        //if selected type: cmd
        "ledCmd": {
          "cmd": "<command>",
          "sec": <interval in seconds [default 5]>
          "blink": <set true to blink and not just on [default false]>
        }
      },
      //examples:
      "3": {
          "type" : "toggle",
          "toggle": {
              "init": false,
              "button": 4
          }
      },
      "4": {
          "type": "cmd",
          "ledCmd": {
              "cmd": "C:\\Users\\mrsag\\git\\ledboard\\scripts\\check_zoom.bat",
              "sec": 5,
              "blink": true
          }
      }
    }
```


<!-- USAGE EXAMPLES -->
## Usage

conf.json is example configuration file for 8 button and leds and it has all of the example uses.
some scripts example can be found in the script folder

1. edit conf.json with the Arduino port in your computer, the button numbers (from your array, starting at 1), and the cmd and led lighting configuration.
2. connect the keyboard.
3. make sure the config file is in the same folder as the script (or use --conf=<path>) and run it:
```
go run .
# or if you pre-compiled on windows:
./ledboard.exe
#on linux:
./ledboard
```


<!-- CONTRIBUTING -->
## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request


<!-- LICENSE -->
## License

Distributed under the MIT License. See `LICENSE` for more information.


<!-- ACKNOWLEDGEMENTS -->
## Acknowledgements
* [Best-README-Template](https://github.com/othneildrew/Best-README-Template)

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->
[contributors-shield]: https://img.shields.io/github/contributors/MRsagi/ledboard.svg?style=for-the-badge
[contributors-url]: https://github.com/MRsagi/ledboard/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/MRsagi/ledboard?style=for-the-badge
[forks-url]: https://github.com/MRsagi/ledboard/network/members
[stars-shield]: https://img.shields.io/github/stars/MRsagi/ledboard.svg?style=for-the-badge
[stars-url]: https://github.com/MRsagi/ledboard/stargazers
[issues-shield]: https://img.shields.io/github/issues/MRsagi/ledboard.svg?style=for-the-badge
[issues-url]: https://github.com/MRsagi/ledboard/issues
[license-shield]: https://img.shields.io/github/license/MRsagi/ledboard.svg?style=for-the-badge
[license-url]: https://github.com/MRsagi/ledboard/blob/master/LICENSE.txt
[linkedin-shield]: https://img.shields.io/badge/-LinkedIn-black.svg?style=for-the-badge&logo=linkedin&colorB=555
[linkedin-url]: https://www.linkedin.com/in/sagi-rosenthal/
