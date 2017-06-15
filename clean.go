// Copyright 2017 strtob01. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	helpAddSyntax           = "Usage: clean add [object]\n\nThe objects are:\n\n\tinteractor\tadd interactor e.g. Order\n\tusecase\tadd usecase e.g. AddItem\n\nUse \"clean help add [object]\" for more information about an object.\n\n"
	helpAddUsecaseSyntax    = "Usage: clean add usecase [usecase] to [interactor]\n\n\tusecase\tname of usecase e.g. AddItem\n\tinteractor\tname of interactor e.g. Order\n\n"
	helpAddInteractorSyntax = "Usage: clean add interactor [name]\n\n\tname\tname of interactor e.g. Order\n\n"
	helpInitSyntax          = "Usage: Use \"clean init\" to initialise a new project in the current folder, i.e. to generate the required boilerplate folders and files. It also sets the Clean Work Directory to the folder in which this command is used.\n\n"
	helpUsage               = "Clean is a tool for generating Clean Architecture boilerplate code.\n\nUsage:\n\n\tclean [verb]\n\nThe verbs are:\n\n\tadd\tadd e.g. new usecase\n\tinit\tinitialise a new Clean Architecture project. Warning! Generates files and folders\n\tset\tset current working directory\n\nUse \"clean help [verb]\" for more information about a verb.\n\nCreated by Tobias Strandberg.\n\n"
	helpSetSyntax           = "Usage: clean set folder\n\nSet the Clean Work Directory to your current directory. The Clean Work Directory is used by Clean to determine in which folders on the hard drive to add interactors and usecases when using e.g. the \"clean add\" command\n\n"
	relPathController       = "ifadapter/controller/"
	relPathPresenter        = "ifadapter/presenter/"
	relPathView             = "ifadapter/view/"
	relPathViewModel        = "ifadapter/view/viewmodel/"
	relPathInteractor       = "usecase/interactor/"
	relPathReqModel         = "usecase/reqmodel/"
	relPathValidator        = "usecase/reqmodel/validator/"
	relPathRespModel        = "usecase/respmodel/"
	verbAdd                 = "add"
	verbInit                = "init"
	verbSet                 = "set"
	verbHelp                = "help"
	objInteractor           = "interactor"
	objUsecase              = "usecase"
	objController           = "controller"
	objView                 = "view"
	objPresenter            = "presenter"
	objValidator            = "validator"
)

var (
	relPaths              = []string{relPathController, relPathPresenter, relPathView, relPathViewModel, relPathInteractor, relPathReqModel, relPathValidator, relPathRespModel}
	projectBaseImportPath string
)

func main() {
	// Sets description for this tool
	flag.Usage = func() {
		fmt.Printf(helpUsage)
	}
	flag.Parse()

	args := flag.Args()
	nArgs := len(args)
	if nArgs == 0 {
		fmt.Printf(helpUsage)
		return
	}
	verb := args[0]
	if verb == verbHelp {
		if nArgs == 1 {
			fmt.Printf(helpUsage)
			return
		}
		switch args[1] {
		case verbAdd:
			if nArgs == 2 {
				fmt.Printf(helpAddSyntax)
			} else if nArgs == 3 {
				switch args[2] {
				case objInteractor:
					fmt.Printf(helpAddInteractorSyntax)
				case objUsecase:
					fmt.Printf(helpAddUsecaseSyntax)
				default:
					fmt.Printf("Invalid object entered.\n\nUse \"clean help add\" for more information about valid objects.\n\n")
				}
			} else {
				fmt.Printf("Invalid number of arguments entered.\n\nUse \"clean help add\" for more information.\n\n")
			}
		case verbInit:
			if nArgs == 2 {
				fmt.Printf(helpInitSyntax)
			} else {
				fmt.Printf("Invalid number of arguments entered.\n\nUse \"clean help init\" for more information\n\n")
			}
		case verbSet:
			if nArgs == 2 {
				fmt.Printf(helpSetSyntax)
			} else {
				fmt.Printf("Invalid number of arguments entered.\n\nUse \"clean help set\" for more information\n\n")
			}
		default:
			fmt.Printf("No such verb, call \"clean -h\" for a list of available verbs.\n\n")
		}
		return
	}

	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Error getting current user: %s\n", err.Error())
		return
	}
	confDir := usr.HomeDir + "/" + ".clean"
	confPath := confDir + "/" + "cleanrc"
	confBytes, err := ioutil.ReadFile(filepath.FromSlash(confPath))
	if err != nil {
		fmt.Printf("Error reading configuration file. Maybe you haven't created a new Clean Architecture Project by executing 'clean init' yet?\n")
		return
	}
	keyValuePairs := bytes.Split(confBytes, []byte("="))
	if keyValuePairs == nil || len(keyValuePairs) < 2 {
		return
	}
	baseDir := string(keyValuePairs[1])
	// Removes the LF character at the end of the string
	baseDir = strings.TrimRight(baseDir, "\n")

	// Find the first occurrence of 'src' and then assume the import path for the project is what follows after that
	// e.g. if baseDir is /users/john/go/src/myproject/ then projectBaseImportPath should be myproject
	found := false
	for i := len(baseDir) - 1; i > 0; i-- {
		if baseDir[i] == 'c' {
			if i > 1 {
				if baseDir[i-1] == 'r' && baseDir[i-2] == 's' {
					projectBaseImportPath = string(baseDir[i+2:])
					found = true
					break
				}
			}
		}
	}
	if !found {
		fmt.Printf("Clean Work Directory not configured. Please go to your project folder and either run \"clean init\" or \"clean set folder\"\n\n")
		return
	}

	// clean [verb]
	switch verb {
	case verbInit:
		if nArgs > 1 {
			fmt.Printf("Invalid number of arguments entered.\n\nUse \"clean help init\" for more information\n\n")
			return
		}
		initProject(filepath.FromSlash(confDir), filepath.FromSlash(confPath))
		return
	case verbSet:
		// User entered: clean set
		if nArgs == 1 {
			fmt.Printf(helpSetSyntax)
			return
		}
		if nArgs == 2 {
			// User entered: clean set jibberish
			if args[1] != "folder" {
				fmt.Printf(helpSetSyntax)
				return
			}
			// User entered: clean set folder
			wd, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error determining current working directory\n")
				return
			}

			// Check for configuration file
			if fileExists(filepath.FromSlash(confPath)) {
				if err := ioutil.WriteFile(
					filepath.FromSlash(confPath),
					[]byte("directory="+filepath.FromSlash(wd)+"/"),
					0700,
				); err != nil {
					fmt.Printf("Error creating config file: %s\n", err.Error())
					return
				}
			} else {
				fmt.Printf("No configuration file exists. Use \"clean init\" to initialise a project instead\n\n")
			}
			return
		}
		fmt.Printf(helpSetSyntax)
		return
	case verbAdd:
		// User entered: clean add
		if nArgs == 1 {
			fmt.Printf(helpAddSyntax)
		} else if nArgs == 2 {
			// User entered: clean add [object]
			switch args[1] {
			case objInteractor:
				// User entered: clean add interactor
				fmt.Printf(helpAddInteractorSyntax)
			case objUsecase:
				// User entered: clean add usecase
				fmt.Printf(helpAddUsecaseSyntax)
			default:
				// User entered: clean add jibberish
				fmt.Printf("Invalid object entered.\n\nUse \"clean help add\" for more information about valid objects.\n\n")
			}
		} else if nArgs == 3 {
			// User entered: clean add [object]
			switch args[1] {
			case objInteractor:
				// User entered: clean add interactor [name]
				ext := filepath.Ext(args[2])
				interactor := string(args[2][:len(args[2])-len(ext)])
				dir := baseDir + "clean/" + relPathController
				addObjToProject(dir, objController, interactor, true)
				dir = baseDir + "clean/" + relPathPresenter
				addObjToProject(dir, objPresenter, interactor, true)
				dir = baseDir + "clean/" + relPathView
				addObjToProject(dir, objView, interactor, true)
				dir = baseDir + "clean/" + relPathInteractor
				addObjToProject(dir, objInteractor, interactor, true)
				dir = baseDir + "clean/" + relPathValidator
				addObjToProject(dir, objValidator, interactor, true)
			case objUsecase:
				// User entered: clean add usecase [usecase]
				fmt.Printf(helpAddUsecaseSyntax)
			default:
				// User entered: clean add jibberish1 jibberish2
				fmt.Printf("Invalid object entered.\n\nUse \"clean help add\" for more information about valid objects.\n\n")
			}
		} else if nArgs == 4 {
			// User entered: clean add [object]
			switch args[1] {
			case objInteractor:
				// User entered: clean add interactor jibberish1 jibberish2
				fmt.Printf("Invalid number of arguments entered.\n\nUse \"clean help add interactor\" for more information.\n\n")
			case objUsecase:
				// User entered: clean add usecase [usecase] to
				fmt.Printf(helpAddUsecaseSyntax)
			default:
				// User entered: clean add jibberish1 jibberish2 jibberish3
				fmt.Printf("Invalid object entered.\n\nUse \"clean help add\" for more information about valid objects.\n\n")
			}
		} else if nArgs == 5 {
			// User entered: clean add [object]
			switch args[1] {
			case objInteractor:
				// User entered: clean add interactor jibberish1 jibberish2 jibberish3
				fmt.Printf("Invalid number of arguments entered.\n\nUse \"clean help add interactor\" for more information.\n\n")
			case objUsecase:
				// User entered: clean add usecase [usecase] to [interactor]
				if args[3] == "to" || args[3] == "To" || args[3] == "tO" || args[3] == "TO" {
					// Remove .go file extension from Object argument
					ext := filepath.Ext(args[4])
					interactor := string(args[4][:len(args[4])-len(ext)])
					for _, v := range relPaths {
						addUsecaseToObject(baseDir+"clean/", v, args[2], interactor)
					}
				} else {
					// User entered: clean add usecase [usecase] jibberish [interactor]
					fmt.Printf(helpAddUsecaseSyntax)
				}
			default:
				// User entered: clean add jibberish1 jibberish2 jibberish3 jibberish4
				fmt.Printf("Invalid object entered.\n\nUse \"clean help add\" for more information about valid objects.\n\n")
			}
		} else {
			fmt.Printf("Invalid number of arguments entered.\n\nUse \"clean help add\" for more information.\n\n")
		}
		return
	default:
		//fmt.Printf("Invalid arguments supplied\n\n")
		fmt.Printf(helpUsage)
	}

}

func addObjToProject(dir, objType, objName string, hasTestFolder bool) {
	// TODO: Remove filename from function signature. The Filename should be the objName + .go
	ext := filepath.Ext(objName)
	var withoutExtFn string
	if ext != ".go" {
		ext = ".go"
		withoutExtFn = objName
	} else {
		withoutExtFn = string(objName[:len(objName)-len(ext)])
	}

	fp := filepath.FromSlash(dir + withoutExtFn + ext)
	if !fileExists(fp) {
		c := fmt.Sprintf("// Package %s provides ... \npackage %s", objType, objType)
		if err := writeBytesToFile(fp, c); err != nil {
			return
		}
		switch objType {
		case objController:
			imports := "\n\nimport (\n\t\"%sclean/usecase/interactor\"\n\t\"%sclean/usecase/reqmodel\"\n)"
			if err := writeBytesToFile(fp, fmt.Sprintf(imports, projectBaseImportPath, projectBaseImportPath)); err != nil {
				return
			}
		case objPresenter:
			imports := "\n\nimport (\n\t\"%sclean/ifadapter/view\"\n\t\"%sclean/ifadapter/view/viewmodel\"\n\t\"%sclean/usecase/respmodel\"\n)"
			if err := writeBytesToFile(fp, fmt.Sprintf(imports, projectBaseImportPath, projectBaseImportPath, projectBaseImportPath)); err != nil {
				return
			}
		case objView:
			imports := "\n\nimport (\n\t\"%sclean/ifadapter/view/viewmodel\"\n)"
			if err := writeBytesToFile(fp, fmt.Sprintf(imports, projectBaseImportPath)); err != nil {
				return
			}
		case objInteractor:
			imports := "\n\nimport (\n\t\"%sclean/ifadapter/presenter\"\n\t\"%sclean/usecase/reqmodel\"\n\t\"%sclean/usecase/reqmodel/validator\"\n\t\"%sclean/usecase/respmodel\"\n)"
			if err := writeBytesToFile(fp, fmt.Sprintf(imports, projectBaseImportPath, projectBaseImportPath, projectBaseImportPath, projectBaseImportPath)); err != nil {
				return
			}
		case objValidator:
			imports := "\n\nimport (\n\t\"%sclean/usecase/reqmodel\"\n\t\"%sclean/usecase/respmodel\"\n)"
			if err := writeBytesToFile(fp, fmt.Sprintf(imports, projectBaseImportPath, projectBaseImportPath)); err != nil {
				return
			}
		}

	}

	// Lower case first character
	lcObjName := firstCharToLower(objName)
	// Upper case first character
	ucObjName := firstCharToUpper(objName)
	// Upper case first character
	ucObjType := firstCharToUpper(objType)

	contentTmpl := "\n\n// %s is a Clean Architecture %s object that wraps its related methods.\n// TODO: Add description of what the interface does\ntype %s interface {\n\t// TODO define interface methods\n}\n\n// %s is an implementation of %s.\ntype %s struct {\n\t// TODO define struct fields and implement the interface\n}"
	content := fmt.Sprintf(contentTmpl, ucObjName, ucObjType, ucObjName, lcObjName, ucObjName, lcObjName)
	if err := writeBytesToFile(fp, content); err != nil {
		return
	}

	if !hasTestFolder {
		return
	}

	testFp := filepath.FromSlash(dir + "test/" + withoutExtFn + "_test" + ext)
	if !fileExists(testFp) {
		c := "// Package test provides ...\npackage test\n\n"
		if err := writeBytesToFile(testFp, c); err != nil {
			return
		}
		if err := writeBytesToFile(testFp, "// TODO: Add tests"); err != nil {
			return
		}
	}

}

// addUsecaseToObject does multiple things.
//  + Adds both a RequestModel and ResponseModel by name of usecaseName
//  + Adds a ViewModel by name of usecaseName
//  + Adds a method by name usecaseName to Controller interface and implementation
//  + Adds a method by name usecaseName to Presenter interface and implementation
//  + Adds a method by name usecaseName to View interface and implementation
//  + Adds a method by name usecaseName to Interactor interface and implementation
//  + Adds a method by name usecaseName to Request Model Validator interface and implementation
func addUsecaseToObject(basePath, relPath, usecaseName, objectName string) {
	fp := filepath.FromSlash(basePath + relPath + firstCharToLower(objectName) + ".go")
	// Check if Object file exists
	fileExists := false
	if _, err := os.Stat(fp); err == nil {
		fileExists = true
	}
	parentDirName := dirNameFromRelPath(relPath)

	if relPath == relPathReqModel || relPath == relPathRespModel || relPath == relPathViewModel {
		// Check if Object file exists, otherwise return
		if _, err := os.Stat(filepath.FromSlash(basePath + relPathPresenter + firstCharToLower(objectName) + ".go")); os.IsNotExist(err) {
			return
		}

		var contentTmpl string
		if !fileExists {
			contentTmpl = fmt.Sprintf("// Package %s provides ...\npackage %s\n", parentDirName, parentDirName)
		} else {
			// Check if struct already exists and return if true
			fileBytes, err := ioutil.ReadFile(fp)
			if err != nil {
				fmt.Printf("Error reading %s: %s\n", fp, err.Error())
				return
			}
			if ix := bytes.Index(fileBytes, []byte(fmt.Sprintf("type %s struct {", firstCharToUpper(usecaseName)))); ix != -1 {
				fmt.Printf("Object already exists\n")
				return
			}
		}
		switch relPath {
		case relPathReqModel:
			contentTmpl = fmt.Sprintf("%s\n// TODO: Add a description.\n// A Clean Architecture RequestModel is a specific usecase's input. More specifically it's the only input argument for the Interactor method which constitutes the usecase.\ntype %s struct {\n\t// TODO: Add struct members\n}", contentTmpl, firstCharToUpper(usecaseName))

		case relPathRespModel:
			contentTmpl = fmt.Sprintf("%s\n// TODO: Add a description.\n// A Clean Architecture ResponseModel is a usecase's specific output. It's used as input to a Presenter method and normally there are more than one ResponseModel corresponding to the same usecase. During the call to the Interactor method all kinds of errors might arise. RequestModel validation errors, authorisation errors and database errors are examples of such outcomes which will all probably require their own ResponseModel.\ntype %s struct {\n\t// TODO: Add struct members\n}\n\n// TODO: Add a description\ntype %sErrVal struct {\n\t// TODO: Add struct members\n}", contentTmpl, firstCharToUpper(usecaseName), firstCharToUpper(usecaseName))

		case relPathViewModel:
			contentTmpl = fmt.Sprintf("%s\n// TODO: Add a description.\n// A Clean Architecture ViewModel is a Presenter's output. It's used as input to a View method and normally there are more than one ViewModel corresponding to the same usecase to accommodate all outcomes such as validation errors, authorisation errors and database errors in addition to the expected usecase outcome.\ntype %s struct {\n\t// TODO: Add struct members\n}\n\n// TODO: Add a description\ntype %sErrVal struct {\n\t// TODO: Add struct members\n}", contentTmpl, firstCharToUpper(usecaseName), firstCharToUpper(usecaseName))
		}
		if err := writeBytesToFile(fp, contentTmpl); err != nil {
			fmt.Printf("Error writing content to reqmodel file: %s\n", err.Error())
		}
		return
	}

	if !fileExists {
		fmt.Printf("Error cannot find the Object file: %s\n\n", fp)
		return
	}

	//fmt.Printf("\n\nProcessing %s\n", fp)
	fileBytes, err := ioutil.ReadFile(fp)
	if err != nil {
		fmt.Printf("Error reading %s: %s\n", fp, err.Error())
		return
	}

	ucObjName := firstCharToUpper(objectName)
	var newFileBytes []byte
	switch relPath {
	case relPathController:
		v := firstCharToUpper(usecaseName)
		// Skip to next fi in the loop in case method already exists
		if ix := bytes.Index(fileBytes, []byte(fmt.Sprintf("%s(", v))); ix != -1 {
			return
		}

		methodSignature := fmt.Sprintf("\t// %s converts the usecase input from its implementation specific format to a Clean Architecture RequestModel, which it then calls the Interactor with.\n\t// TODO: Add description\n\t%s()\n", v, v)
		newFileBytes, err = addMethodSignatureToInterface(fileBytes, fp, methodSignature, ucObjName)
		if err != nil {
			fmt.Printf("Error in addMethodSignatureToInterface: %s\n", err.Error())
			return
		}
		method := fmt.Sprintf("\n\n// %s implements the %s interface method %s.\nfunc (%s *%s) %s() {\n\t// TODO: Implement interface method\n}", v, firstCharToUpper(objectName), v, firstCharInWord(firstCharToLower(objectName)), firstCharToLower(objectName), v)
		newFileBytes, err = addMethodToImpl(newFileBytes, method, objectName)
		if err != nil {
			fmt.Printf("Error in addMethodToImpl: %s\n", err.Error())
			return
		}
	case relPathPresenter:
		v := firstCharToUpper(usecaseName)
		// Skip to next fi in the loop in case method already exists
		if ix := bytes.Index(fileBytes, []byte(fmt.Sprintf("Present%s(", v))); ix != -1 {
			return
		}

		methodSignature := fmt.Sprintf("\t// Present%s converts the usecase output i.e. a ResponseModel to a Clean Architecture ViewModel, which it then calls the View with.\n\t// TODO: Add description\n\tPresent%s(rsm *respmodel.%s)\n\t// Present%sErrVal converts the validation failure ResponseModel to a corresponding ViewModel.\n\t// TODO: Add description\n\tPresent%sErrVal(rsm *respmodel.%sErrVal)\n", v, v, v, v, v, v)
		newFileBytes, err = addMethodSignatureToInterface(fileBytes, fp, methodSignature, ucObjName)
		if err != nil {
			fmt.Printf("Error in addMethodSignatureToInterface: %s\n", err.Error())
			return
		}
		method := fmt.Sprintf("\n\n// Present%s implements the %s interface method Present%s.\nfunc (%s *%s) Present%s(rsm *respmodel.%s) {\n\t// TODO: Implement interface method\n}\n// Present%sErrVal implements the %s interface method Present%sErrVal.\nfunc (%s *%s) Present%sErrVal(rsm *respmodel.%sErrVal) {\n\t// TODO: Implement interface method\n}", v, firstCharToUpper(objectName), v, firstCharInWord(firstCharToLower(objectName)), firstCharToLower(objectName), v, v, v, firstCharToUpper(objectName), v, firstCharInWord(firstCharToLower(objectName)), firstCharToLower(objectName), v, v)
		newFileBytes, err = addMethodToImpl(newFileBytes, method, objectName)
		if err != nil {
			fmt.Printf("Error in addMethodToImpl: %s\n", err.Error())
			return
		}
	case relPathView:
		v := firstCharToUpper(usecaseName)
		// Skip to next fi in the loop in case method already exists
		if ix := bytes.Index(fileBytes, []byte(fmt.Sprintf("Render%s(", v))); ix != -1 {
			return
		}

		methodSignature := fmt.Sprintf("\t// Render%s renders the View in an application specific format. It builds the View exclusively from the ViewModel.\n\t// TODO: Add description\n\tRender%s(vm *viewmodel.%s)\n\t// Render%sErrVal renders the validation failure View in an application specific format. It builds the View exclusively from the ViewModel.\n\t// TODO: Add description\n\tRender%sErrVal(vm *viewmodel.%sErrVal)\n", v, v, v, v, v, v)
		newFileBytes, err = addMethodSignatureToInterface(fileBytes, fp, methodSignature, ucObjName)
		if err != nil {
			fmt.Printf("Error in addMethodSignatureToInterface: %s\n", err.Error())
			return
		}
		method := fmt.Sprintf("\n\n// Render%s implements the %s interface method Render%s.\nfunc (%s *%s) Render%s(vm *viewmodel.%s) {\n\t// TODO: Implement interface method\n}\n\n// Render%sErrVal implements the %s interface method Render%sErrVal.\nfunc (%s *%s) Render%sErrVal(vm *viewmodel.%sErrVal) {\n\t// TODO: Implement interface method\n}", v, firstCharToUpper(objectName), v, firstCharInWord(firstCharToLower(objectName)), firstCharToLower(objectName), v, v, v, firstCharToUpper(objectName), v, firstCharInWord(firstCharToLower(objectName)), firstCharToLower(objectName), v, v)
		newFileBytes, err = addMethodToImpl(newFileBytes, method, objectName)
		if err != nil {
			fmt.Printf("Error in addMethodToImpl: %s\n", err.Error())
			return
		}
	case relPathInteractor:
		v := firstCharToUpper(usecaseName)
		// Skip to next fi in the loop in case method already exists
		if ix := bytes.Index(fileBytes, []byte(fmt.Sprintf("%s(", v))); ix != -1 {
			return
		}

		methodSignature := fmt.Sprintf("\t// %s is a Clean Architecture Interactor method whose input is a RequestModel. It orchestrates all of the steps that together constitute the usecase such as RequestModel validation, calling of Gateways for manipulation of databases and more. Given the outcome of the steps taken, it assembles a specific ResponseModel and calls a specific Presenter method with the ResponseModel as input.\n\t// TODO: Add description.\n\t%s(rqm *reqmodel.%s)\n", v, v, v)
		newFileBytes, err = addMethodSignatureToInterface(fileBytes, fp, methodSignature, ucObjName)
		if err != nil {
			fmt.Printf("Error in addMethodSignatureToInterface: %s\n", err.Error())
			return
		}
		method := fmt.Sprintf("\n\n// %s implements the %s interface method %s.\nfunc (%s *%s) %s(rqm *reqmodel.%s) {\n\t// TODO: Implement interface method\n}", v, firstCharToUpper(objectName), v, firstCharInWord(firstCharToLower(objectName)), firstCharToLower(objectName), v, v)
		newFileBytes, err = addMethodToImpl(newFileBytes, method, objectName)
		if err != nil {
			fmt.Printf("Error in addMethodToImpl: %s\n", err.Error())
			return
		}
	case relPathValidator:
		v := firstCharToUpper(usecaseName)
		// Skip to next fi in the loop in case method already exists
		if ix := bytes.Index(fileBytes, []byte(fmt.Sprintf("Validate%s(", v))); ix != -1 {
			return
		}

		methodSignature := fmt.Sprintf("\t// Validate%s validates rqm. If valid it returns nil otherwise an %sErrVal\n\tValidate%s(rqm *reqmodel.%s) *respmodel.%sErrVal\n", v, v, v, v, v)
		newFileBytes, err = addMethodSignatureToInterface(fileBytes, fp, methodSignature, ucObjName)
		if err != nil {
			fmt.Printf("Error in addMethodSignatureToInterface: %s\n", err.Error())
			return
		}
		method := fmt.Sprintf("\n\n// Validate%s implements the %s interface method Validate%s.\nfunc (%s *%s) Validate%s(rqm *reqmodel.%s) *respmodel.%sErrVal {\n\t// TODO: Implement interface method\n\treturn nil\n}", v, firstCharToUpper(objectName), v, firstCharInWord(firstCharToLower(objectName)), firstCharToLower(objectName), v, v, v)
		newFileBytes, err = addMethodToImpl(newFileBytes, method, objectName)
		if err != nil {
			fmt.Printf("Error in addMethodToImpl: %s\n", err.Error())
			return
		}
	}
	if err := ioutil.WriteFile(fp, newFileBytes, 0700); err != nil {
		fmt.Printf("Error writing to %s: %s\n", fp, err.Error())
		return
	}
}

func dirNameFromRelPath(relPath string) string {
	// relPath is of the form 'ifadapter/controller/'
	pieces := strings.Split(relPath, "/")
	if pieces == nil {
		return ""
	}
	return pieces[len(pieces)-2]
}

// addMethodSignatureToInterface adds a method to the interface ifName
func addMethodSignatureToInterface(b []byte, filepath, methodSignature, ifName string) ([]byte, error) {
	pieces := bytes.SplitAfter(b, []byte(fmt.Sprintf("type %s interface {\n", ifName)))
	if len(pieces) != 2 {
		fmt.Printf("%s content not split into two halves\n", filepath)
		return nil, errors.New("Error splitting b")
	}
	methodSignatureBytes := []byte(methodSignature)
	p1Reader := bytes.NewReader(pieces[0])
	var w bytes.Buffer
	if _, err := io.Copy(&w, p1Reader); err != nil {
		fmt.Printf("Error copying from p1Reader to w: %s\n", err.Error())
		return nil, err
	}
	usecaseReader := bytes.NewReader(methodSignatureBytes)
	if _, err := io.Copy(&w, usecaseReader); err != nil {
		fmt.Printf("Error copying from usecaseReader to w: %s\n", err.Error())
		return nil, err
	}
	p2Reader := bytes.NewReader(pieces[1])
	if _, err := io.Copy(&w, p2Reader); err != nil {
		fmt.Printf("Error copying from p2Reader to w: %s\n", err.Error())
		return nil, err
	}
	return w.Bytes(), nil
}

func addMethodToImpl(b []byte, method, implName string) ([]byte, error) {
	startingIx := bytes.Index(b, []byte(fmt.Sprintf("type %s struct {\n", firstCharToLower(implName))))
	if startingIx == -1 {
		fmt.Printf("Implementation %s not found\n", implName)
		return nil, errors.New("Implementation not found")
	}
	leftBracesN := 0
	rightBracesN := 0
	found := false
	implClosingBracketIx := 0
	for i := startingIx; i < len(b); i++ {
		if b[i] == '{' {
			leftBracesN++
			continue
		}
		if b[i] == '}' {
			rightBracesN++
			if rightBracesN == leftBracesN && rightBracesN > 0 {
				implClosingBracketIx = i
				found = true
				break
			}
		}
	}
	if !found {
		fmt.Printf("Couldn't find the implementation\n")
		return nil, errors.New("Couldn't find the implementation")
	}
	newbuf := make([]byte, len(b)+len(method))
	for k, v := range b[:implClosingBracketIx+1] {
		newbuf[k] = v
	}
	offset := implClosingBracketIx + 1
	for i := 0; i < len(method); i++ {
		newbuf[offset+i] = method[i]
	}
	offset = implClosingBracketIx + len(method) + 1
	for k, v := range b[implClosingBracketIx+1:] {
		newbuf[offset+k] = v
	}
	return newbuf, nil

	// Check if the following syntax is possible
	// for _, v := range b[startingIx:] {
	// }
}

func fileExists(filepath string) bool {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		// path to confPath does not exist
		return false
	}
	return true
}

func writeBytesToFile(filepath string, content string) error {
	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0700)
	defer f.Close()
	if err != nil {
		fmt.Printf("Error opening file: %s\n", err.Error())
		return err
	}
	if _, err = f.WriteString(content); err != nil {
		fmt.Printf("Error writing to file: %s\n", err.Error())
		return err
	}
	return nil
}

// firstCharInWord returns the first character in word
func firstCharInWord(word string) string {
	// Lower case first character
	for _, v := range word {
		return string(v)
	}
	return ""
}

// firstCharToLower returns text after lowering its first character
func firstCharToLower(text string) string {
	// Lower case first character
	var output string
	for _, v := range text {
		s := string(v)
		ls := strings.ToLower(s)
		output = ls + text[len(s):]
		break
	}
	return output
}

// firstCharToUpper returns text after capitalising its first character
func firstCharToUpper(text string) string {
	var output string
	for _, v := range text {
		s := string(v)
		us := strings.ToUpper(s)
		output = us + text[len(s):]
		break
	}
	return output
}

func initProject(confDir, confPath string) {
	fmt.Printf("Adding folder structure to current directory...\n")
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error determining current working directory\n")
		return
	}

	// Check for configuration file
	if !fileExists(confPath) {
		// path to confPath does not exist
		if !mkdir(confDir) {
			return
		}
		if err := ioutil.WriteFile(
			confPath,
			[]byte("directory="+filepath.FromSlash(wd)+"/"),
			0700,
		); err != nil {
			fmt.Printf("Error creating config file: %s\n", err.Error())
			return
		}
	} else {
		if err := ioutil.WriteFile(
			confPath,
			[]byte("directory="+filepath.FromSlash(wd)+"/"),
			0700,
		); err != nil {
			fmt.Printf("Error creating config file: %s\n", err.Error())
			return
		}
	}

	if !mkdir("clean") {
		return
	}
	if !mkdir("clean/entity") {
		return
	}
	if !mkdir("clean/ifadapter") {
		return
	}
	if !mkdir("clean/ifadapter/controller") {
		return
	}
	if !mkdir("clean/ifadapter/controller/test") {
		return
	}
	if !mkdir("clean/ifadapter/gateway") {
		return
	}
	if !mkdir("clean/ifadapter/gateway/test") {
		return
	}
	if !mkdir("clean/ifadapter/presenter") {
		return
	}
	if !mkdir("clean/ifadapter/presenter/test") {
		return
	}
	if !mkdir("clean/ifadapter/view") {
		return
	}
	if !mkdir("clean/ifadapter/view/test") {
		return
	}
	if !mkdir("clean/ifadapter/view/viewmodel") {
		return
	}
	if !mkdir("clean/usecase") {
		return
	}
	if !mkdir("clean/usecase/interactor") {
		return
	}
	if !mkdir("clean/usecase/interactor/test") {
		return
	}
	if !mkdir("clean/usecase/reqmodel") {
		return
	}
	if !mkdir("clean/usecase/reqmodel/validator") {
		return
	}
	if !mkdir("clean/usecase/reqmodel/validator/test") {
		return
	}
	if !mkdir("clean/usecase/respmodel") {
		return
	}

	if !mkdir("lib") {
		return
	}
	if !mkdir("cmd") {
		return
	}
	//fmt.Printf("Base Directory: %s\n", filepath.Base(ex))
}

func mkdir(name string) bool {
	if err := os.Mkdir(name, 0700); err != nil {
		fmt.Printf("Error creating the folder '%s': %s\n", name, err.Error())
		return false
	}
	return true
}

// clean init
// ==========
// Generates necessary files and folders for a new Clean Architecture project
//
// clean add interactor AppInstance
// clean remove interactor AppInstance
// clean add gateway AppInstance
// clean add entity Order
// clean add
