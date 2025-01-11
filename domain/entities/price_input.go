package entities


// calculo inicial alpha
// PriceInput representa os dados necessários para o "Cálculo Inicial Alpha".

type PriceInput struct{
	// tabela productscmp
	IcmsMedio float64
	PisCofinsMedio float64
	CustoMedioLiq float64
	CustoMedioNF float64

	// outros módulos
	IcmsEfetivo float64
	IcmsEfetivoPR float64
	IcmsMinas float64
	IcmsTriangular float64
	Difal float64
}

// CalculationDetails contém os detalhes do cálculo
type CalculationDetails struct {
	SKU           string  `json:"sku"`
	IcmsEfetivo   float64 `json:"icms_efetivo"`
	Difal         float64 `json:"difal"`
	IcmsMedioCalc float64 `json:"icms_medio_calc"`
	PisCofinsCalc float64 `json:"pis_cofins_calc"`
	CustoMedio    float64 `json:"custo_medio"`
	Operacao      float64 `json:"operacao"`
	Comissao      float64 `json:"comissao"`
	LucroPadrao   float64 `json:"lucro_padrao"`
	Fcp           float64 `json:"fcp"`
	Imposto       float64 `json:"imposto"`
	ValorFinal    float64 `json:"valor_final"`
}
