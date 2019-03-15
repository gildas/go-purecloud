package purecloud

// commit contains the current git commit and is set in the build.sh script
var commit string

// VERSION is the version of this application
var VERSION = "0.0.1" + commit

// APP is the name of the application
const APP string = "PureCloud Agent"
