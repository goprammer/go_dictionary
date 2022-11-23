## About
This is dictionary shows how to efficiently upload and serve large text files in Go. The spell-checker uses character manipulation and prefix matching to provide a best guess with decent performance.

The Gutenburg_Dictionary directory holds a csv version of the public domain portion of "The Project Gutenberg Etext of Webster's Unabridged Dictionary" which is in turn based on the 1913 US Webster's Unabridged Dictionary.

The Gutenburg_EBook directory holds the Project Gutenberg EBook of The Adventures of Sherlock Holmes. This is used to determine the probability of encountering a given word in everyday speech, which aids the spell-checker. 

Assuming you have permission, a newer dictionary and a larger sample of published material would improve the relevancy of the definitions and the spelling corrections.

## Run as CLI
`go run cmd_line.go`

Within CLI, search for definition of 'dog'
`	> dog`
The dictionary will give you every definition it has for 'dog'.

Within CLI, search for definition of 'dogz'
`	> dogz`
The dictionary will inform you that it couldn't find the word 'dogz' and that it's using 'dog' instead.

Within CLI, pass the -d flag
`	> dogz -d`
If the word is misspelled or not present in the dictionay, you will receive a debug message from the spell-checker explanationing why it chose a different word.

Gracefully shutdown
`ctrl+C`

## Run as a RESTful API with a swagger page
Load a port number and public domain.
`source ./ENV.sh`

Start server
`go run REST_API.go`

visit the swagger page to see the expected JSON input.
`localhost:8888/docs`

Gracefully shutdown. This catches the delete command from K8s.
`kill -15 <pid>`

