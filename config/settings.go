package config

// ForbiddenFileAndFolderNames Folders and files for which calculations are carried out will not be
var ForbiddenFileAndFolderNames = []string{
	"settings",
	"config",
}

// ApprovedExtensions File extensions for which the calculation will be carried out
var ApprovedExtensions = []string{
	".go",
}

// ProjectsPaths Path to the folder where git is initialized
var ProjectsPaths = []string{
	"/var/www/test/project1",
	"/var/www/test/project2",
}
