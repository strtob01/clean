# clean
[Clean Architecture](https://8thlight.com/blog/uncle-bob/2012/08/13/the-clean-architecture.html) Boilerplate Generation Tool for Go
## About
Clean generates files, folders and Go code according to a standard based on the Clean Architecture design pattern.

Clean Architecture has its merits but it also implies creating structs, interfaces and methods in abundance, which can be daunting even for the most tenacious. Clean helps you generate the backbone of a Clean Architecture application, enabling you to focus on the important parts and less on how to structure you code. All this while adhering to the Clean Architecture design pattern.
## Prerequisites
1. The [Go](https://golang.org) Programming language installed on your computer
## Installation
1. Use `go get github.com/strtob01/clean` to download and install the Clean package onto your computer.
2. Go to the Clean folder `cd "$GOPATH/src/github.com/strtob01/clean"`
3. Compile the Clean Project `go build`
4. Move the generated clean binary to any of your bin folders e.g. `mv clean "$GOPATH/bin"`
5. Verify by typing `clean` which should execute the clean application and you should see a textual output "Clean is a tool for..."
## Usage
When starting on a new application Clean can generate a good starting point in terms of package structure.
Assume a developer has decided to create a new package called example in the $GOPATH/src folder so that the path to the package is $GOPATH/src/example. To use Clean the user enters the example folder by entering `cd "$GOPATH/src/example"`. Then the user uses `clean init` to generate the basic folders. Now the example folder contains a tree of new folders as in the table below.

Folder | Description
--- | ---
clean | The clean folder is the root of your Clean Architecture folder structure
clean/entity | The entity folder holds all of your Entities
clean/ifadapter | The ifadapter folder is the root of your Interface Adapters
clean/ifadapter/controller | The controller folder contains all of your Controllers
clean/ifadapter/controller/test | The test folder is where you write unit tests for the Controllers
clean/ifadapter/gateway | The gateway folder contains all of your Gateways
clean/ifadapter/gateway/test | The test folder is where you write unit tests for the Gateways
clean/ifadapter/presenter | The presenter folder contains all of your Presenters
clean/ifadapter/presenter/test | The test folder is where you write unit tests for the Presenters
clean/ifadapter/view | The view folder contains all of your Views
clean/ifadapter/view/test | The test folder is where you write unit tests for the Views
clean/ifadapter/view/viewmodel | The viewmodel folder contains all of your ViewModels
clean/usecase | The usecase folder is the root of your Usecases
clean/usecase/interactor | The interactor folder contains all of your Interactors
clean/usecase/interactor/test | The test folder is where you write unit tests for the Interactors
clean/usecase/reqmodel | The reqmodel folder contains all of your RequestModels
clean/usecase/reqmodel/validator | The validator folder is where your write validators for your RequestModels
clean/usecase/reqmodel/validator/test | The test folder is where you write unit tests for the Validators
clean/usecase/respmodel | The respmodel folder contains all of your ResponseModels
cmd | The cmd folder is where you keep files containing the main() function. For example, if you want the binary file generated when running `go build` to be named tripplanner you would create a folder called tripplanner inside the cmd folder e.g. `mkdir tripplanner` and finally creating a Go file containing the func main() and placing it inside the tripplanner folder.
lib | The lib folder contains all project specific libraries that you create or download from the Internet

The `clean init` command saves the path to the "$GOPATH/src/example" folder in a hidden file on your hard drive. Any subsequent commands such as `clean add ...` use this path to create and generate Go code in the correct files. If you're working on several projects and need to switch to a different one then go to the root folder of the other project e.g. `cd "$GOPATH/src/anotherProject"` and run `clean set folder`. This updates the hidden file with the new working directory.

To properly use Clean it's necessary to understand the roles of the controllers, gateways, presenters etc. An example of a Controller is given below:
```Go
type Example interface {
  AddItemToOrder(r *http.Request)
}
type example struct {
  ia interactor.Example
}
func (e *example) AddItemToOrder(r *http.Request) {
  // Add implementation here...
}
```
In the table below these roles are described:

Object | Calls | Called by | Role Description
--- | --- | --- | ---
Controller | Interactor method | A Http.Handler function in the case of a webserver app | Controller is a really bad naming since its role is more about converting external input to something the Clean Architecture application understands. In short, a Controller's sole responsibility is to convert external input such as data in a http.Request data structure to a RequestModel and finally call the Interactor's Usecase method with the RequestModel as input argument. A Controller is therefore the glue between the Clean and non-Clean part of your code.
Interactor | Gateways, Validator, Presenter | Controller | Constitutes the Usecase and orchestrates all of the actions required by the Usecase such as RequestModel validation, getting/setting data from/to Gateways, assembling a ResponseModel and more. It always ends with the Interactor calling a Presenter method with the assembled ResponseModel as input argument. Which Presenter method depends on the outcome of the actions in the Usecase. A perfect example would be a validation failure of the RequestModel, which is the only input argument in the Interactor method. The Interactor should then call a Presenter method which knows how to present the failed validation to the user of the application.
Validator | Nothing | Interactor | A Validator is not a Clean Architecture object, but it has a role to play in the flow of a Clean Architecture application written using the Clean tool. A Validator is in fact a RequestModel validator. Interactors should validate RequestModels and having a separate object to perform this validation enables unit tests for the validation. Validators take the Controller supplied RequestModel and validate it and return nil if it validates, otherwise a ValidationFailedResponseModel.
Presenter | View | Interactor | A Presenter is the equivalent of the Controller but for the output of the Clean application. Its sole responsibility is to convert the input argument, the ResponseModel, to a ViewModel and call the appropriate View method with the assembled ViewModel as input argument. The ViewModel should be a very simple data structure with only struct fields and no methods. The ViewModel should contain all the data required by the View to be able to render the UI/response. In the case of a web page form where the user has entered some incorrect data, the RequestModel validation should fail. The ViewModel should not only contain fields for all the fields of the form, but also booleans flags and such so that the View can correctly render the UI to the user which form fields were not correctly filled out.
View | A http.ResponseWriter in the case of a webserver app | Presenter | A View's sole responsibility is to render the output of the application in an application specific way. It translates the ViewModel into e.g. a HTML response in the case of a webserver.
Gateway | E.g. databases, external APIs etc | Interactor | A Gateway is responsible for being an interface between your application and something external such as a database or external API.

Before going any further you need to think about which Usecases there are for your application. The `clean add ...` command uses two definitions, usecase and interactor.

Object | Description
--- | ---
usecase | A usecase is a specific action that your application performs. In an order handling system there would be usecases such as AddItemToOrder, RemoveItemFromOrder etc.
interactor | Think of an interactor as an object that groups related usecases together.

After deciding on which Usecases your application should have. Do an analysis of which of them are related to each other and then think of a fitting name for each group of related Usecases. Each group of related Usecases corresponds to an Interactor object in your code. In the example of an order handling system such an Interactor could be named OrderHandler for the AddItemToOrder and RemoveItemFromOrder Usecases.

The Clean tool requires that you add your Interactors before adding their Usecases. In the case of the beforementioned example of an order handling system you would now run the `clean add interactor OrderHandler` command.
This creates a string of files all containing an interface called OrderHandler and an implementation in the form of a struct called orderHandler. Notice that the interface implementation is defined with a lower case first letter. That is because your other packages should only know about interfaces for testability. Each call to `clean add interactor ...` creates Go files with the same name as the `...`. The idea is that each file should only contain one group of related Usecases. A file named `...` is created in the controller, presenter, view, validator and interactor folders. It also creates a file called `..._test.go` in each of the test folders. After calling `clean add interactor OrderHandler` the content of the file created in the interactor folder would be:
```Go
package interactor

import (
	"path/to/your/project/clean/ifadapter/presenter"
	"path/to/your/project/clean/usecase/reqmodel"
	"path/to/your/project/clean/usecase/reqmodel/validator"
	"path/to/your/project/clean/usecase/respmodel"
)

type OrderHandler interface {
}
type orderHandler struct {
	val	validator.OrderHandler
	ps 	presenter.OrderHandler
}
```
The next step is to add the Usecases to the OrderHandler. This is done by using the `clean add usecase AddItemToOrder to OrderHandler` and the `clean add usecase RemoveItemFromOrder to OrderHandler` command. This would update all of the files in the controller, presenter, view, validator and interactor folders. Our example file above would now look like:
```Go
package interactor

import (
	"path/to/your/project/clean/ifadapter/presenter"
	"path/to/your/project/clean/usecase/reqmodel"
	"path/to/your/project/clean/usecase/reqmodel/validator"
	"path/to/your/project/clean/usecase/respmodel"
)

type OrderHandler interface {
	AddItemToOrder(rqm *reqmodel.AddItemToOrder)
	RemoveItemFromOrder(rqm *reqmodel.RemoveItemFromOrder)
}
type orderHandler struct {
	val	validator.OrderHandler
	ps 	presenter.OrderHandler
}
func (o *orderHandler) AddItemToOrder(rqm *reqmodel.AddItemToOrder) {
	// Request Model Validation
	if rsm := o.val.ValidateAddItemToOrder(rqm); rsm != nil {
		o.ps.PresentAddItemToOrderErrVal(rsm)
		return
	}
}
func (o *orderHandler) RemoveItemFromOrder(rqm *reqmodel.RemoveItemFromOrder) {
	// Request Model Validation
	if rsm := o.val.ValidateRemoveItemFromOrder(rqm); rsm != nil {
		o.ps.PresentRemoveItemFromOrderErrVal(rsm)
		return
	}
}
```
In addition to this the above commands also generate additional code. Only the output of the former command is shown below though:
1. An AddItemToOrder RequestModel.
2. An AddItemToOrder and AddItemToOrderErrVal ResponseModels. The latter is used by the Presenter to handle the scenario where the supplied RequestModel fails validation.
3. An AddItemToOrder and AddItemToOrderErrVal ViewModels.

And that's pretty much all there's to the Clean tool. I hope you find it useful.
## Contributing
1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request :D
## History
###### June 15th, 2017
Project initated
## Credits
Tobias Strandberg
## License
[The MIT License](https://opensource.org/licenses/MIT)
