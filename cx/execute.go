package cxcore

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/SkycoinProject/skycoin/src/cipher/encoder"
)

// It "un-runs" a program
// func (prgrm *CXProgram) Reset() {
// 	prgrm.CallStack = MakeCallStack(0)
// 	prgrm.Steps = make([][]CXCall, 0)
// 	prgrm.Outputs = make([]*CXArgument, 0)
// 	//prgrm.ProgramSteps = nil
// }

// UnRun ...
func (cxt *CXProgram) UnRun(nCalls int) {
	if nCalls >= 0 || cxt.CallCounter < 0 {
		return
	}

	call := &cxt.CallStack[cxt.CallCounter]

	for c := nCalls; c < 0; c++ {
		if call.Line >= c {
			// then we stay in this call counter
			call.Line += c
			c -= c
		} else {

			if cxt.CallCounter == 0 {
				call.Line = 0
				return
			}
			c += call.Line
			call.Line = 0
			cxt.CallCounter--
			call = &cxt.CallStack[cxt.CallCounter]
		}
	}
}

// ToCall ...
func (cxt *CXProgram) ToCall() *CXExpression {
	for c := cxt.CallCounter - 1; c >= 0; c-- {
		if cxt.CallStack[c].Line+1 >= len(cxt.CallStack[c].Operator.Expressions) {
			// then it'll also return from this function call; continue
			continue
		}
		return cxt.CallStack[c].Operator.Expressions[cxt.CallStack[c].Line+1]
		// prgrm.CallStack[c].Operator.Expressions[prgrm.CallStack[prgrm.CallCounter-1].Line + 1]
	}
	// error
	return &CXExpression{Operator: MakeFunction("", "", -1)}
	// panic("")
}

// func (cxt *CXProgram) GetCall()

// Run ...
func (cxt *CXProgram) Run(untilEnd bool, nCalls *int, untilCall int) error {
	defer RuntimeError()
	var err error

	// // Checking if expression is a goroutine.
	// // Getting expression that will be executed.
	// expr := cxt.GetExpr()
	// if expr.IsGoRoutine {
	// 	thread := cxt.Threads[cxt.ThreadCount]
	// 	thread.Terminated = false
	// 	cxt.ThreadCount++

	// 	thread.StackFloor, thread.StackCeiling = AllocateStack(THREAD_STACK_SIZE)
	// 	// It's a new thread (no data has been
	// 	// written), so `StackPointer` is the same as the `StackFloor`
	// 	thread.StackPointer = thread.StackFloor

	// 	thread.CallFloor, thread.CallCeiling = AllocateCallStack(THREAD_CALLSTACK_SIZE)
	// }

	for !cxt.Terminated && (untilEnd || *nCalls != 0) && cxt.CallCounter > untilCall {
		// call := &cxt.CallStack[cxt.CallCounter]
		call := cxt.GetCall(true)

		// checking if enough memory in stack
		if cxt.StackPointer > STACK_SIZE {
			panic(STACK_OVERFLOW_ERROR)
		}

		if !untilEnd {
			var inName string
			var toCallName string
			var toCall *CXExpression

			if call.Line >= call.Operator.Length && cxt.CallCounter == 0 {
				cxt.Terminated = true
				cxt.CallStack[0].Operator = nil
				cxt.CallCounter = 0
				fmt.Println("in:terminated")
				return err
			}

			if call.Line >= call.Operator.Length && cxt.CallCounter != 0 {
				toCall = cxt.ToCall()
				// toCall = prgrm.CallStack[prgrm.CallCounter-1].Operator.Expressions[prgrm.CallStack[prgrm.CallCounter-1].Line + 1]
				inName = cxt.CallStack[cxt.CallCounter-1].Operator.Name
			} else {
				toCall = call.Operator.Expressions[call.Line]
				inName = call.Operator.Name
			}

			if toCall.Operator == nil {
				// then it's a declaration
				toCallName = "declaration"
			} else if toCall.Operator.IsNative {
				toCallName = OpNames[toCall.Operator.OpCode]
			} else {
				if toCall.Operator.Name != "" {
					toCallName = toCall.Operator.Package.Name + "." + toCall.Operator.Name
				} else {
					// then it's the end of the program got from nested function calls
					cxt.Terminated = true
					cxt.CallStack[0].Operator = nil
					cxt.CallCounter = 0
					fmt.Println("in:terminated")
					return err
				}
			}

			fmt.Printf("in:%s, expr#:%d, calling:%s()\n", inName, call.Line+1, toCallName)
			*nCalls--
		}

		err = call.ccall(cxt)
		if err != nil {
			return err
		}
	}

	return nil
}

// minHeapSize determines what's the minimum heap size that a CX program
// needs to have based on INIT_HEAP_SIZE, MAX_HEAP_SIZE and NULL_HEAP_ADDRESS_OFFSET.
func minHeapSize() int {
	minHeapSize := INIT_HEAP_SIZE
	if MAX_HEAP_SIZE < INIT_HEAP_SIZE {
		// Then MAX_HEAP_SIZE overrides INIT_HEAP_SIZE's value.
		minHeapSize = MAX_HEAP_SIZE
	}
	if minHeapSize < THREAD_STACK_SIZE {
		// Then the user is trying to allocate too little heap memory.
		// We need at least THREAD_STACK_SIZE bytes for `nil`.
		minHeapSize = THREAD_STACK_SIZE
	}

	return minHeapSize
}

// EnsureHeap ensures that `cxt` has `minHeapSize()`
// bytes allocated after the data segment.
func (cxt *CXProgram) EnsureHeap() {
	currHeapSize := len(cxt.Memory) - cxt.HeapStartsAt
	minHeapSize := minHeapSize()
	if currHeapSize < minHeapSize {
		cxt.Memory = append(cxt.Memory, make([]byte, minHeapSize-currHeapSize)...)
	}
}

// RunCompiled ...
func (cxt *CXProgram) RunCompiled(nCalls int, args []string) error {
	_, err := cxt.SelectProgram()
	if err != nil {
		panic(err)
	}
	cxt.EnsureHeap()
	rand.Seed(time.Now().UTC().UnixNano())

	var untilEnd bool
	if nCalls == 0 {
		untilEnd = true
	}
	mod, err := cxt.SelectPackage(MAIN_PKG)
	if err != nil {
		return err
	}

	if cxt.CallStack[0].Operator == nil {
		// then the program is just starting and we need to run the SYS_INIT_FUNC
		fn, err := mod.SelectFunction(SYS_INIT_FUNC)
		if err != nil {
			return err
		}
		// *init function
		mainCall := MakeCall(fn)
		cxt.CallStack[0] = mainCall
		cxt.StackPointer = fn.Size

		for !cxt.Terminated {
			call := &cxt.CallStack[cxt.CallCounter]
			err := call.ccall(cxt)
			if err != nil {
				return err
			}
		}
		// we reset call state
		cxt.Terminated = false
		cxt.CallCounter = 0
		cxt.CallStack[0].Operator = nil
	}

	fn, err := mod.SelectFunction(MAIN_FUNC)
	if err != nil {
		return err
	}

	if len(fn.Expressions) < 1 {
		return nil
	}

	if cxt.CallStack[0].Operator == nil {
		// main function
		mainCall := MakeCall(fn)
		mainCall.FramePointer = cxt.StackPointer
		// initializing program resources
		cxt.CallStack[0] = mainCall

		// prgrm.Stacks = append(prgrm.Stacks, MakeStack(1024))
		cxt.StackPointer += fn.Size

		// feeding os.Args
		osPkg, err := PROGRAM.SelectPackage(OS_PKG)
		if err == nil {
			argsOffset := 0

			osGbl, err := osPkg.GetGlobal(OS_ARGS)
			if err != nil {
				return err
			}
			
			for _, arg := range args {
				argBytes := encoder.Serialize(arg)
				argOffset := AllocateSeq(len(argBytes) + OBJECT_HEADER_SIZE)

				var header = make([]byte, OBJECT_HEADER_SIZE)
				WriteMemI32(header, 5, int32(encoder.Size(arg)+OBJECT_HEADER_SIZE))
				obj := append(header, argBytes...)

				WriteMemory(argOffset, obj)

				var argOffsetBytes [4]byte
				WriteMemI32(argOffsetBytes[:], 0, int32(argOffset))
				argsOffset = WriteToSlice(argsOffset, argOffsetBytes[:])
			}
			WriteI32(GetFinalOffset(0, osGbl), int32(argsOffset))
		}

		cxt.Terminated = false
	}

	if err = cxt.Run(untilEnd, &nCalls, -1); err != nil {
		return err
	}

	if cxt.Terminated {
		cxt.Terminated = false
		cxt.CallCounter = 0
		cxt.CallStack[0].Operator = nil
	}

	return nil
}

func (call *CXCall) ccall(prgrm *CXProgram) error {
	if call.Line >= call.Operator.Length {
		/*
		   popping the stack
		*/
		// going back to the previous call
		var callCounter *int
		callCounter = prgrm.GetCallCounter()
		*callCounter--
		callFloor := prgrm.GetThreadCallFloor()

		if *callCounter < callFloor {
			// then the program finished
			thread := prgrm.GetThread()
			if thread != nil {
				thread.Terminated = true
				prgrm.CompactThreads()
			} else {
				prgrm.Terminated = true
			}
		} else {
			// copying the outputs to the previous stack frame
			if prgrm.GetThread() != nil {
				return nil
			}
			returnAddr := &prgrm.CallStack[*callCounter]
			returnOp := returnAddr.Operator
			returnLine := returnAddr.Line
			returnFP := returnAddr.FramePointer
			fp := call.FramePointer

			expr := returnOp.Expressions[returnLine]

			// lenOuts := len(expr.Outputs)
			for i, out := range call.Operator.Outputs {
				WriteMemory(
					GetFinalOffset(returnFP, expr.Outputs[i]),
					ReadMemory(
						GetFinalOffset(fp, out),
						out))
			}

			// return the stack pointer to its previous state
			prgrm.StackPointer = call.FramePointer
			// we'll now execute the next command
			prgrm.CallStack[*callCounter].Line++
			// calling the actual command
			// prgrm.CallStack[prgrm.CallCounter].ccall(prgrm)
		}
	} else {
		/*
		   continue with call operator's execution
		*/

		// Checking if expression is a goroutine.
		// Current expression being executed.
		expr := PROGRAM.GetExpr()

		// fn := call.Operator
		// expr := fn.Expressions[call.Line]
		// if it's a native, then we just process the arguments with execNative
		if expr.Operator == nil {
			// then it's a declaration
			// wiping this declaration's memory (removing garbage)
			var callCounter *int
			callCounter = prgrm.GetCallCounter()
			newCall := &prgrm.CallStack[*callCounter]
			newFP := newCall.FramePointer
			size := GetSize(expr.Outputs[0])
			for c := 0; c < size; c++ {
				prgrm.Memory[newFP+expr.Outputs[0].Offset+c] = 0
			}
			call.Line++
		} else if expr.Operator.IsNative {
			// go func() {
			// 	execNative(prgrm)
			// }()
			// call.Line++

			execNative(prgrm)
			call.Line++
			// prgrm.AdvanceThread()
			// if PROGRAM.ThreadCount > 0 && PROGRAM.ThreadCounter < 0 {
			// 	PROGRAM.ThreadCounter++
			// 	// _ = prgrm.GetCall(true)
			// }
		} else {
			/*
			   It was not a native, so we need to create another call
			   with the current expression's operator
			*/
			// we're going to use the next call in the callstack
			// prgrm.CallCounter++
			var callCounter *int
			callCounter = prgrm.GetCallCounter()
			
			if *callCounter >= CALLSTACK_SIZE {
				panic(STACK_OVERFLOW_ERROR)
			}

			if expr.IsGoRoutine {
				thread := &PROGRAM.Threads[PROGRAM.ThreadCount]
				thread.Terminated = false
				PROGRAM.ThreadCount++
				if PROGRAM.ThreadCounter < 0 {
					PROGRAM.ThreadCounter++
				}
				// PROGRAM.ThreadCounter++

				thread.StackFloor, thread.StackCeiling = AllocateStack(THREAD_STACK_SIZE)
				// It's a new thread (no data has been
				// written), so `StackPointer` is the same as the `StackFloor`
				thread.StackPointer = thread.StackFloor

				thread.CallFloor, thread.CallCeiling = AllocateCallStack(THREAD_CALLSTACK_SIZE)
				thread.CallCounter = thread.CallFloor

				
				callCounter = prgrm.GetCallCounter()
				// *callCounter++
				call.Line++
			} else {
				*callCounter++
			}

			newCall := &prgrm.CallStack[*callCounter]
			// setting the new call
			newCall.Operator = expr.Operator
			newCall.Line = 0
			newCall.FramePointer = prgrm.StackPointer
			// the stack pointer is moved to create room for the next call
			// prgrm.MemoryPointer += fn.Size
			prgrm.StackPointer += newCall.Operator.Size

			// checking if enough memory in stack
			if prgrm.StackPointer > prgrm.StackCeiling {
				// Trying to expand current thread's stack.
				// prgrm.ExpandStack()
			}

			fp := call.FramePointer
			newFP := newCall.FramePointer

			// wiping next stack frame (removing garbage)
			for c := 0; c < expr.Operator.Size; c++ {
				prgrm.Memory[newFP+c] = 0
			}

			for i, inp := range expr.Inputs {
				var byts []byte
				// finalOffset := inp.Offset
				finalOffset := GetFinalOffset(fp, inp)
				// finalOffset := fp + inp.Offset

				// if inp.Indexes != nil {
				// 	finalOffset = GetFinalOffset(&prgrm.Stacks[0], fp, inp)
				// }
				if inp.PassBy == PASSBY_REFERENCE {
					// If we're referencing an inner element, like an element of a slice (&slc[0])
					// or a field of a struct (&struct.fld) we no longer need to add
					// the OBJECT_HEADER_SIZE to the offset
					if inp.IsInnerReference {
						finalOffset -= OBJECT_HEADER_SIZE
					}
					var finalOffsetB [4]byte
					WriteMemI32(finalOffsetB[:], 0, int32(finalOffset))
					byts = finalOffsetB[:]
				} else {
					size := GetSize(inp)
					byts = prgrm.Memory[finalOffset : finalOffset+size]
				}

				// writing inputs to new stack frame
				WriteMemory(
					GetFinalOffset(newFP, newCall.Operator.Inputs[i]),
					// newFP + newCall.Operator.Inputs[i].Offset,
					// GetFinalOffset(prgrm.Memory, newFP, newCall.Operator.Inputs[i], MEM_WRITE),
					byts)
			}
		}
	}
	return nil
}

// Callback ...
func (cxt *CXProgram) Callback(expr *CXExpression, functionName string, packageName string, inputs [][]byte) {
	if fn, err := cxt.GetFunction(functionName, packageName); err == nil {
		line := cxt.CallStack[cxt.CallCounter].Line
		previousCall := cxt.CallCounter
		cxt.CallCounter++
		newCall := &cxt.CallStack[cxt.CallCounter]
		newCall.Operator = fn
		newCall.Line = 0
		newCall.FramePointer = cxt.StackPointer
		cxt.StackPointer += newCall.Operator.Size
		newFP := newCall.FramePointer

		// wiping next mem frame (removing garbage)
		for c := 0; c < expr.Operator.Size; c++ {
			cxt.Memory[newFP+c] = 0
		}

		for i, inp := range inputs {
			WriteMemory(GetFinalOffset(newFP, newCall.Operator.Inputs[i]), inp)
		}

		var nCalls = 0
		if err := cxt.Run(true, &nCalls, previousCall); err != nil {
			os.Exit(CX_INTERNAL_ERROR)
		}

		cxt.CallCounter = previousCall
		cxt.CallStack[cxt.CallCounter].Line = line
	}
}
