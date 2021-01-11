package parsnip

type operatorSend struct {
	value                     evaluatableValue
	oppositeSideInsideBracket bool
	oppositeSideOperator      byte
}

type variable struct {
	value evaluatableValue

	insideBracketRight bool
	insideBracketLeft  bool
	rightOperator      byte
	leftOperator       byte
	rightSend          chan operatorSend
	leftSend           chan operatorSend
}
