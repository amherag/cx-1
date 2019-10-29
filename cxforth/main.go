package main

import (
	"log"
	"os"
	"strconv"
	"strings"
	"bufio"
	"flag"
	"fmt"

	cx "github.com/SkycoinProject/cx/cx"
	// parser "github.com/SkycoinProject/cx/cxgo/parser"
	// actions "github.com/SkycoinProject/cx/cxgo/actions"
)

var (
	in *bufio.Reader
	DataStack []float64
	PROGRAM *cx.CXProgram
	// Used to indicate the parser if a new function is being defined.
	// isDefFunc
)

var file *string = flag.String("file", "", "Define a CXForth src file to be read.");

func main() {
	flag.Parse()

	PROGRAM = cx.MakeProgram()
	mainPkg := cx.MakePackage(cx.MAIN_PKG)
	PROGRAM.AddPackage(mainPkg)
	mainFn := cx.MakeFunction(cx.MAIN_FUNC, "", 0)
	mainPkg.AddFunction(mainFn)

	initFn := cx.MakeFunction(cx.SYS_INIT_FUNC, "", 0)
	mainPkg.AddFunction(initFn)

	// parser.FunctionDeclaration(initFn, nil, nil, nil)
	PROGRAM.SelectFunction(cx.MAIN_FUNC)

	if *file != "" {
		file_to_read, err := os.OpenFile(*file, os.O_RDONLY, 0755); 
		if file_to_read == nil { log.Fatal(err); }
		
		in = bufio.NewReader( file_to_read );
		
		comment := false; //if we're currently in a comment
		
		for {
			dat, err := in.ReadString(' ');
			if err != nil { log.Fatal("can't read file string"); return; }
			
			if dat[0:len(dat)-1] == "(" { 
				comment = true;
			}
			
			if comment == false && dat[0:len(dat)-1] != "" && dat[0:len(dat)-1] != " " {
				parse_forth(dat[0:len(dat)-1]);
			}
			
			if dat[0:len(dat)-1] == ")"  {
				comment = false;
			}
		}
	} else {
		buf := bufio.NewReader(os.Stdin)

		for {
			fmt.Print("> ");
			read, err := buf.ReadString('\n')
			if err != nil {
				println()
				break
			}
			
			comms := strings.Split(read, " ")
			
			for i := 0; i < len(comms); i++ {
				if len(comms) == 0 || comms[i] == "" || comms[i] == " " {
					continue
				}

				parse_forth(strings.TrimSpace(comms[i]))
			}
		}
	}
}
/*
	Experimental, session type operation (think python shell)
	
	
	else {
		in = bufio.NewReader(os.Stdin);
		for {
			dat, err := in.ReadString(' ');
			if err != nil { log.Stderr(err); return; }
			
			parse_forth(dat[0:len(dat)-1], DataStack);
			}
	}
	/**/

// func addInitFunction (prgrm *cx.CXProgram) {
// 	if main, err := prgrm.GetPackage(cx.MAIN_PKG); err == nil {
// 		initFn := cx.MakeFunction(cx.SYS_INIT_FUNC, "", 0)
// 		main.AddFunction(initFn)

// 		// parser.FunctionDeclaration(initFn, nil, nil, nil)
// 		prgrm.SelectFunction(cx.MAIN_FUNC)
// 	} else {
// 		panic(err)
// 	}
// }

func check_stack_size(required int) bool {
	if len(DataStack) < required {
		log.Fatal("Stack depth is less then " + string(required) + ". Operation is impossible");
		return false;
	} 
	
	return true;
}

func parse_forth(dat string) {
	switch strings.TrimSpace(string(dat)) {
	case "":
	case "<cr>":
		return;
	case "t":
		//check the DataStack size using the popped value
		//	if it passes, then the program continues
		minimum, DataStack := DataStack[len(DataStack)-1], DataStack[:len(DataStack)-1]
		// minimum := int(DataStack.Pop().(float64));
		if len(DataStack) < int(minimum) {
			log.Println("DataStack has not enough minerals (values)");
		}
	case ":dp":
		PROGRAM.PrintProgram()
	case ".":
		var x float64
		x, DataStack = DataStack[len(DataStack)-1], DataStack[:len(DataStack)-1]
		log.Println(x);
	// case ":":
		
	// case "0SP":
	// 	DataStack.Cut(0, L);
	// case ".S":
	// 	log.Println(DataStack);
	// case "2/":
	// 	DataStack.Push( DataStack.Pop().(float64) / 2);
	// case "2*":
	// 	DataStack.Push( DataStack.Pop().(float64) * 2);
	// case "2-":
	// 	DataStack.Push( DataStack.Pop().(float64) - 2);
	// case "2+":
	// 	DataStack.Push( DataStack.Pop().(float64) + 2);
	// case "1-":
	// 	DataStack.Push( DataStack.Pop().(float64) - 1);
	// case "1+":
	// 	DataStack.Push( DataStack.Pop().(float64) + 1);
	// case "DUP":
	// 	DataStack.Push( DataStack.Last() );
	// case "?DUP":
	// 	if DataStack.Last().(float64) != 0 { DataStack.Push( DataStack.Last().(float64) ); }
	// case "PICK":
	// 	number := int(DataStack.Pop().(float64)) ;
		
	// 	if number < L {
	// 		DataStack.Push( DataStack.At(L - 1 - number).(float64) );
	// 	} else {
	// 		log.Fatal("picking out of stack not allowed. Stack Length: " + string(L) + ". Selecting: " + string(number) + ".");
	// 		return;
	// 	}
	// case "TUCK":
	// 	DataStack.Insert(L - 2, int(DataStack.Last().(float64)) );
	// case "NIP":
	// 	DataStack.Delete(L - 2);
	// case "2DROP":
	// 	DataStack.Pop(); DataStack.Pop();
	// case "2DUP":
	// 	DataStack.Push(DataStack.At(L - 2));
	// 	DataStack.Push(DataStack.At(DataStack.Len() - 2));
	// case "DROP":
	// 	DataStack.Pop();
	// case "OVER":
	// 	DataStack.Push(DataStack.At(L - 2));
	// case "SWAP":
	// 	l := DataStack.Len();
	// 	DataStack.Swap(l - 2, l - 1);
	// case "*":
	// 	num1 := DataStack.Pop().(float64);
	// 	num2 := DataStack.Pop().(float64);
	// 	DataStack.Push( num1 * num2 );				
	// case "+":
		// num1 := DataStack.Pop().(float64);
		// num2 := DataStack.Pop().(float64);
		// DataStack.Push( num1 + num2 );
	// case "-":
	// 	num1 := DataStack.Pop().(float64);
	// 	num2 := DataStack.Pop().(float64);
	// 	DataStack.Push( num2 - num1 );
	// case "/":
	// 	num1 := DataStack.Pop().(float64);
	// 	num2 := DataStack.Pop().(float64);
	// 	DataStack.Push( num2 / num1 );
	// case "-ROT":
	// 	DataStack.Swap(L - 1, L - 2);
	// 	DataStack.Swap(L - 2, L - 3);
	// case "ROT":
	// 	DataStack.Swap(L - 3, L - 2);
	// 	DataStack.Swap(L - 2, L - 1);
	// case "2OVER":
	// 	DataStack.Push(DataStack.At(L - 4));
	// 	DataStack.Push(DataStack.At(DataStack.Len() - 4));
	// case "2SWAP":
	// 	DataStack.Swap(L - 4, L - 2);
	// 	DataStack.Swap(L - 3, L - 1);
	// case "EMIT":
	// 	log.Println( string([]byte{uint8(DataStack.Last().(float64))}) );
	default:
		val, err := strconv.ParseFloat(dat, 64);
		
		if err == nil {
			DataStack = append(DataStack, val)
			PROGRAM.Memory
		} else {
			log.Println(err);
			log.Fatalln("error, unknown token \""+dat+"\"");
		}
	}
}
