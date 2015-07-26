package main

import (
    "bufio"
    "fmt"
    "io"
    "io/ioutil"
    "log"
    "os"
    "path/filepath"
    "regexp"
    "strings"
)

// The name of the british american word list file
var britishAmerican = "british-american.txt"

// When starting the program, prepend the directory name to the name of
// word list file
func init() {
    dir, _ := filepath.Split(os.Args[0])
    britishAmerican = filepath.Join(dir, britishAmerican)
}

func main() {
    inFilename, outFilename, err := filenamesFromCommandLine()
    if err != nil {
        // Print usage information contained in err and exit
        fmt.Println(err)
        os.Exit(1)
    }
    
    inFile, outFile := os.Stdin, os.Stdout
    
    if inFilename != "" {
        if inFile, err = os.Open(inFilename); err != nil {
            log.Fatal(err)
        }
        
        defer inFile.Close()
    }
    
    if outFilename != "" {
        if outFile, err = os.Create(outFilename); err != nil {
            log.Fatal(err)
        }
        
        defer outFile.Close()
    }
    
    if err := americanise(inFile, outFile); err != nil {
        log.Fatal(err)
    }
}

// Read the command line arguments
func filenamesFromCommandLine() (inFilename, outFilename string, err error) {
    if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
        // User has asked for help: return usage string
        err = fmt.Errorf("usage: %s [<]infile.txt [>]outfile.txt",
            filepath.Base(os.Args[0]))
        return "", "", err
    }
    
    // If there are arguments, they must be file names
    if len(os.Args) > 1 {
        inFilename = os.Args[1]
        if len(os.Args) > 2 {
            outFilename = os.Args[2]
        }
    }
    
    // Sanity check: in and out file should be different
    if inFilename != "" && inFilename == outFilename {
        log.Fatal("won't overwrite the infile")
    }
    
    return inFilename, outFilename, nil
}

// Replace british english words from input with us english words to output
func americanise(inFile io.Reader, outFile io.Writer) (err error) {
    reader := bufio.NewReader(inFile)
    writer := bufio.NewWriter(outFile)
    
    // Flush the out file at the end of function
    defer func() {
        if err == nil {
            err = writer.Flush()
        }
    }()
    
    // The replacer function used by the regexp replacement method
    var replacer func(string) string
    if replacer, err = makeReplacerFunction(britishAmerican); err != nil {
        return err
    }
    
    // Use regular expression for applying the replacer function
    wordRx := regexp.MustCompile("[A-Za-z]+")
    
    eof := false
    for !eof {
        var line string
        
        line, err = reader.ReadString('\n')
        if err != nil {
            if err == io.EOF {
                // EOF is not really an error, but it's time to exit the loop
                // at the next iteration
                err = nil
                eof = true
            } else {
                return err
            }
        }
        
        // Here, finally the replacer does its magic
        line = wordRx.ReplaceAllStringFunc(line, replacer)
        
        // Write the us english line to output
        if _, err = writer.WriteString(line); err != nil {
            return err
        }
    }
    
    // We got that far without any error...
    return nil
}

func makeReplacerFunction(file string) (func(string) string, error) {
    // Read the bytes of the file in one go
    rawBytes, err := ioutil.ReadFile(file)
    if err != nil {
        return nil, err
    }
    
    // Convert the utf-8 bytes into an utf-8 string
    text := string(rawBytes)
    
    // Make a map (dictionary) for the brit-us dictionary
    usForBritish := make(map[string]string)
    
    // Split the input text into lines and add each key-value pair to the map
    lines := strings.Split(text, "\n")
    for _, line := range lines {
        fields := strings.Fields(line)
        if len(fields) == 2 {
            usForBritish[fields[0]] = fields[1]
        }
    }
    
    // Now return the replacer function
    return func(word string) string {
        if usWord, found := usForBritish[word]; found {
            // There is an us replacement 
            return usWord
        }
        // There is no us replacement
        return word
    }, nil
}