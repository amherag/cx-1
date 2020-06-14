package cxcore

import (
	"net/http"

	. "github.com/SkycoinProject/cx/cx"
	"github.com/SkycoinProject/skycoin/src/cipher/encoder"
	"github.com/SkycoinProject/dmsg"
	"github.com/SkycoinProject/dmsg/disc"
	"github.com/SkycoinProject/dmsg/cipher"

	"github.com/jinzhu/copier"
)

var clients map[[33]byte]*dmsg.Client = make(map[[33]byte]*dmsg.Client, 0)

func init() {
	dmsgPkg := MakePackage("dmsg")
	clientStrct := MakeStruct("Client")

	pubKeyFld := MakeArgument("pubKey", "", 0).AddType(TypeNames[TYPE_UI8]).AddPackage(dmsgPkg)
	pubKeyFld.DeclarationSpecifiers = append(pubKeyFld.DeclarationSpecifiers, DECL_ARRAY)
	pubKeyFld.TotalSize = 33
	pubKeyFld.IsArray = true
	pubKeyFld.Lengths = []int{33}
	clientStrct.AddField(pubKeyFld)

	dmsgPkg.AddStruct(clientStrct)

	PROGRAM.AddPackage(dmsgPkg)
}

func opDMSGNewClient(prgrm *CXProgram) {
	expr := prgrm.GetExpr()
	fp := prgrm.GetFramePointer()
	inp1, inp2, out1 := expr.Inputs[0], expr.Inputs[1], expr.Outputs[0]

	dmsgD := disc.NewMock()

	sPK := ReadArray(fp, inp1, inp1.Type).([]byte)
	sSK := ReadArray(fp, inp2, inp1.Type).([]byte)

	var bPK [33]byte
	for c := 0; c < len(bPK); c++ {
		bPK[c] = byte(sPK[c])
	}

	var bSK [32]byte
	for c := 0; c < len(bSK); c++ {
		bSK[c] = byte(sSK[c])
	}

	dmsgClient := dmsg.NewClient(cipher.PubKey(bPK), cipher.SecKey(bSK), dmsgD, dmsg.DefaultConfig())
	clients[bPK] = dmsgClient

	// Output structure `dmsg.Client`.
	cli := CXArgument{}
	err := copier.Copy(&cli, out1)
	if err != nil {
		panic(err)
	}

	// Extracting CX `dmsg` package.
	dmsgPkg, err := PROGRAM.GetPackage("dmsg")
	if err != nil {
		panic(err)
	}

	// Extracting `dmsg`'s Client structure.
	clientType, err := dmsgPkg.GetStruct("Client")
	if err != nil {
		panic(err)
	}

	// Extracting `dmsg.Client`'s `pubKey` field.
	pubKeyFld, err := clientType.GetField("pubKey")
	if err != nil {
		panic(err)
	}

	accessAddr := []*CXArgument{pubKeyFld}
	cli.Fields = accessAddr
	// WriteString(fp, cliAddr, &cli)
	WriteMemory(GetFinalOffset(fp, &cli), sPK)
}

func opDMSGClientServe(prgrm *CXProgram) {
	expr := prgrm.GetExpr()
	fp := prgrm.GetFramePointer()
	inp1 := expr.Inputs[0]

	// Input structure `Client`.
	cli := CXArgument{}
	err := copier.Copy(&cli, inp1)
	if err != nil {
		panic(err)
	}

	// Extracting CX `dmsg` package.
	dmsgPkg, err := PROGRAM.GetPackage("dmsg")
	if err != nil {
		panic(err)
	}

	// Extracting `dmsg`'s Client structure.
	cliType, err := dmsgPkg.GetStruct("Client")
	if err != nil {
		panic(err)
	}

	// Extracting `dmsg.Client`'s `pubKey` field.
	pubKeyFld, err := cliType.GetField("pubKey")
	if err != nil {
		panic(err)
	}

	// Getting corresponding `Client` instance.
	accessPubKey := []*CXArgument{pubKeyFld}
	cli.Fields = accessPubKey

	sPK := ReadArray(fp, &cli, cli.Type).([]byte)
	
	var bPK [33]byte
	for c := 0; c < len(bPK); c++ {
		bPK[c] = byte(sPK[c])
	}
	
	client := clients[bPK]
	go client.Serve()
}

func opDMSGDo(prgrm *CXProgram) {
	expr := prgrm.GetExpr()
	fp := prgrm.GetFramePointer()

	inp1, out1 := expr.Inputs[0], expr.Outputs[0]
	var req http.Request
	byts1 := ReadMemory(GetFinalOffset(fp, inp1), inp1)
	err := encoder.DeserializeRawExact(byts1, &req)
	if err != nil {
		WriteString(fp, err.Error(), out1)
	}
}
