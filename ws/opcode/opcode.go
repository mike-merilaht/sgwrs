package opcode

type OpCode int
const (
	OpCodeContinuation OpCode = iota
	OpCodeText 
	OpCodeBinary

	OpCodeReserved3
	OpCodeReserved4
	OpCodeReserved5
	OpCodeReserved6
	OpCodeReserved7

	OpCodeClosed

	OpCodePing
	OpCodePong

	OpCodeReserved11
	OpCodeReserved12
	OpCodeReserved13
	OpCodeReserved14
	OpCodeReserved15
)

var OpCodeStrings = []string{
		"0 (Continuation)",
		"1 (Text)",
		"2 (Binary)",
		"3 (RSV3)",
		"4 (RSV4)",
		"5 (RSV5)",
		"6 (RSV6)",
		"7 (RSV7)",
		"8 (Closed)",
		"9 (Ping)",
		"10 (Pong)",
		"11 (RSV11)",
		"12 (RSV12)",
		"13 (RSV13)",
		"14 (RSV14)",
		"15 (RSV15)",
	};

func (opcode OpCode) String() string {
    return OpCodeStrings[opcode]
}
