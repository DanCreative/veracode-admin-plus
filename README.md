# veracode-admin-plus
**Veracode Admin Plus** is a local web application utility. It is intended to make the life of a Veracode administrator easier, by dramatically speeding up the process of updating multiple users' permissions. 

Make changes to one or more users | Optionally view all of the changes | Submit all of the changes
:--:|:--:|:--:
<img src="./docs/assets/changes.gif" height="150"/> | <img src="./docs/assets/view_cart.gif" height="150"/> | <img src="./docs/assets/submit_cart.gif" height="150"/>

## Features
- **It's fast:** Asynchronously fetches and updates users in bulk. (Also its a Go application😉)
- **Filtering:** All filtering available on the official Veracode UI, is available in this utility.

## Getting Started
### Installation
0. Download and install go from: https://go.dev/doc/install
1. Open a terminal and run command: ```go install "github.com/DanCreative/veracode-admin-plus"```

### Configuration
1. **Veracode Admin Plus** makes use of the same Veracode API credential file pattern that the other Veracode utilities use. This is handy, because it means that you can manage all of your API credentials in the same place. Please follow the Veracode documentation to create a new API credentials file if you don't have one already ([Windows](https://docs.veracode.com/r/t_configure_credentials_windows)/[macOS or Linux](https://docs.veracode.com/r/t_configure_credentials_mac)).
2. If you have more than one profile, you can set the ```VERACODE_API_PROFILE``` environment variable to switch between them. This works in the same way as the HTTPie Veracode authentication library's [multipe profiles](https://docs.veracode.com/r/c_httpie_tool#using-multiple-profiles) feature.
3. **Veracode Admin Plus** On the initial execution of the application, pass below configuration values. Example: ```vap -p 8082 -r "us" -s```.
   
Parameter | Default Value | Description
:--:|:--:|:--:
|-r Region | com | Set the region where your organization's Veracode account data is hosted. Possible values: eu, com or us.
|-p Port | 8080 | Set the port on which the utility will run.
|-s Save | false | Passing the -s flag, will save the region and port values to the profile. This can be useful in the case that you have different profiles for different instances of Veracode in different regions. Once the configuration values are saved, you don't have to continue to pass them as command line arguments. 

## Technologies
### Frontend
- Vanilla HTML, CSS and Javascript
- HTMX
- AlpineJS
- JQuery
- Hyperscript
### Backend
- Go
- chi