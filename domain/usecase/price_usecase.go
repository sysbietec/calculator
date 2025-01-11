package usecase

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"
	"calculator/domain/entities"
	"calculator/domain/repositories"
	"calculator/internal/firebird"
)

// PriceUseCase define os métodos do caso de uso de cálculo
type PriceUseCase interface {
	CalculateAlphaPrice(sku string) (float64, string, error)
	CalculateAlphaPriceWithUserPrice(sku string, userPrice float64) (float64, string, error)
}

// priceUseCaseImpl implementa PriceUseCase
type priceUseCaseImpl struct {
	productRepo    repositories.ProductRepository
	productService *firebird.ProductService
}

// NewPriceUseCase "injeta" o repositório para o caso de uso
func NewPriceUseCase(pr repositories.ProductRepository, ps *firebird.ProductService) PriceUseCase {
	return &priceUseCaseImpl{
		productRepo:    pr,
		productService: ps,
	}
}

func (uc *priceUseCaseImpl) CalculateAlphaPrice(sku string) (float64, string, error) {
	return uc.CalculateAlphaPriceWithUserPrice(sku, 0)
}

 
func SimulateProfit(precoDigitado, custoMedio, custoMedioNF, frete, rebate, res1, i9, operacao, comissao, fcp float64) (float64, error) {
	// L1 = custoMedio + (CustoMedioNF * 0.01)
	L1 := custoMedio + (custoMedioNF * 0.01)

	// L2 = frete - (frete * rebate)
	L2 := frete - (frete * rebate)

	// L3 = L1 + L2
	L3 := L1 + L2

	// L4 = preco_digitado * (res1 + i9 + operacao + comissao + fcp)
	L4 := precoDigitado * (res1 + i9 + operacao + comissao + fcp)

	// L5 = L4 + L3
	L5 := L4 + L3

	// L6 = preco_digitado - L5
	L6 := precoDigitado - L5

	// L7 = L6 * 100
	L7 := L6 * 100

	// L8 = (preco_digitado / L7) / 100
	if L7 == 0 {
		return 0, fmt.Errorf("divisão por zero em L7")
	}
	L8 := (precoDigitado / L7) / 100

	return L8, nil
}


// CalculateAlphaPrice é o método que orquestra a busca de dados e executa a fórmula de cálculo
func (uc *priceUseCaseImpl) CalculateAlphaPriceWithUserPrice(sku string, userPrice float64) (float64, string, error) {
	 
	// Converter SKU para int
	produto, err := strconv.Atoi(sku)
	if err != nil {
		return 0, "", fmt.Errorf("erro ao converter SKU para int: %w", err)
	}

	// 1. Buscar do repositório: dados do productscmp → retorna PriceInput (parcial)
	priceInp, err := uc.productRepo.GetProductCmpValues(sku)
	if err != nil {
		return 0, "", fmt.Errorf("erro ao GetProductCmpValues: %w", err)
	}

	// 2. Buscar parâmetros padrão
	params, err := uc.productRepo.GetParameters()
	if err != nil {
		return 0, "", fmt.Errorf("erro ao GetParameters: %w", err)
	}

	// 3. Buscar CostFire (comissão, frete, departamento) no Firebird
	costF, err := uc.productRepo.GetCostFire(sku)
	if err != nil {
		return 0, "", fmt.Errorf("erro ao GetCostFire: %w", err)
	}

	// 4. Consultar o ICMS Efetivo e Difal usando o ProductService
	icmsVenda, difal, err := uc.productService.CalculateIcmsEfetivoEDifal(produto)
	if err != nil {
		return 0, "", fmt.Errorf("erro ao consultar ICMS Efetivo e Difal: %w", err)
	}

	// Ajustar os valores no PriceInput
	priceInp.IcmsEfetivo = icmsVenda
	priceInp.Difal = difal

	logrus.WithFields(logrus.Fields{
		"IcmsEfetivo": priceInp.IcmsEfetivo,
		"Difal":       priceInp.Difal,
	}).Info("ICMS Efetivo e Difal ajustados")

	// Calcular o valor final e capturar as variáveis principais
	valorFinal, lucroSimulado, calcErr := alphaCalculation(priceInp, params, costF, userPrice)
	if calcErr != nil {
		return 0, "", fmt.Errorf("erro alphaCalculation: %w", calcErr)
	}

	// Montar o JSON com as variáveis principais
	calculationDetails := map[string]interface{}{
		"sku":             sku,
		"icms_efetivo":    priceInp.IcmsEfetivo,
		"difal":           priceInp.Difal,
		"icms_medio_calc": priceInp.IcmsMedio * 0.4,
		"pis_cofins_calc": priceInp.PisCofinsMedio * 0.4,
		"custo_medio_liq": priceInp.CustoMedioLiq ,
		"custo_medio_calc":priceInp.CustoMedioLiq + (priceInp.IcmsMedio * 0.4) + (priceInp.PisCofinsMedio * 0.4),
		"operacao":        params.Operacao,
		"comissao":        costF.Comissao,
		"lucro_padrao":    params.LucroPadraoDesejado,
		"fcp":             params.Fcp,
		"frete":           costF.Frete,
		"rebate":          params.Rebate,
		"custo_medio_nf":  priceInp.CustoMedioNF,
		"Preço Tabela U02":     valorFinal,
		"Lucro Simulado":  lucroSimulado, 
	}

	// Serializar o mapa em JSON
	calculationDetailsJSON, err := json.MarshalIndent(calculationDetails, "", "  ")
	if err != nil {
		return 0, "", fmt.Errorf("erro ao serializar detalhes do cálculo em JSON: %w", err)
	}

	// Logar o JSON gerado
	logrus.WithField("calculation_details", string(calculationDetailsJSON)).Info("Detalhes completos do cálculo gerados")

	// Retornar o valor final e o JSON como string
	return valorFinal, string(calculationDetailsJSON), nil
}


// alphaCalculation implementa a lógica do Cálculo Inicial Alpha
func alphaCalculation(pi entities.PriceInput, pm entities.Parameters, cf entities.CostFire, userPrice float64) (float64, float64, error) {
	// Log dos parâmetros recebidos
	logrus.WithFields(logrus.Fields{
		"PriceInput":  pi,
		"Parameters":  pm,
		"CostFire":    cf,
	}).Info("Valores recebidos para cálculo")

	// Cálculo inicial
	icmsMedioCalc := pi.IcmsMedio * 0.4
	pisCofinsCalc := pi.PisCofinsMedio * 0.4
	comissao := cf.Comissao / 100

	logrus.WithFields(logrus.Fields{
		"icmsMedioCalc":    icmsMedioCalc,
		"pisCofinsCalc":    pisCofinsCalc,
	}).Info("Valores calculados de ICMS e Pis/Cofins")

	// Cálculo do custo médio
	custoMedio := pi.CustoMedioLiq + icmsMedioCalc + pisCofinsCalc
	logrus.WithField("custoMedio", custoMedio).Info("Custo médio calculado")

	// Cálculo de i1
	i1 := (pm.AliquotaPis + pm.AliquotaCofins) * pm.RedutorPadrao
	logrus.WithField("i1", i1).Info("Valor de i1 calculado")

	// Cálculo de i2, i3, i4 e res1
	i2 := (pi.IcmsEfetivo - pi.Difal)
	i3 := (i2 + pi.IcmsEfetivo) / 2
	i4 := 1 - i3
	res1 := i4 * i1
	logrus.WithFields(logrus.Fields{
		"i2":   i2,
		"i3":   i3,
		"i4":   i4,
		"res1": res1,
	}).Info("Valores intermediários calculados (i2, i3, i4, res1)")

	// Cálculo de i6, i7, i8, i9
	i6 := pi.IcmsEfetivo * 0.60
	i7 := (pi.IcmsEfetivo - pi.Difal) * 0.60
	i8 := i7 + pi.Difal
	i9 := (i8 + i6) / 2
	logrus.WithFields(logrus.Fields{
		"i6": i6,
		"i7": i7,
		"i8": i8,
		"i9": i9,
	}).Info("Valores intermediários calculados (i6, i7, i8, i9)")

	// Cálculo de res3
	logrus.WithFields(logrus.Fields{
		"Operacao":          pm.Operacao,
		"Comissao":          comissao,
		"LucroPadraoDesejado": pm.LucroPadraoDesejado,
		"Fcp":               pm.Fcp,
	}).Info("Calculando res3")
	res3 := pm.Operacao + comissao + pm.LucroPadraoDesejado + pm.Fcp
	logrus.WithField("res3", res3).Info("Valor de res3 calculado")

	// Cálculo do imposto
	imposto := 1 - (res1 + i9 + res3)
	
	if imposto == 0 {
		return 0, 0, fmt.Errorf("imposto zero => division by zero")
	}

	logrus.WithFields(logrus.Fields{
		"res1": res1,
		"i9":   i9,
		"res3": res3,
	}).Info("Cálculo final do imposto")
	logrus.WithField("imposto", imposto).Info("Imposto calculado")

	// Cálculo final
	logrus.WithFields(logrus.Fields{
		"custoMedio":        custoMedio,
		"CustoMedioNF":      pi.CustoMedioNF,
		"Frete":             cf.Frete,
		"FreteRebate":       cf.Frete,
		"Rebate":	pm.Rebate,
	}).Info("Detalhamento do cálculo de valorFinal")
	
	logrus.WithFields(logrus.Fields{
		"message":"Iniciando agora simulador de lucro",
	}).Info()

	simulatorProfit, err:= SimulateProfit(userPrice, custoMedio, pi.CustoMedioNF, cf.Frete, pm.Rebate, res1, i9, pm.Operacao, comissao, pm.Fcp) 
	if err != nil {
		return 0 ,0 ,fmt.Errorf("erro ao calcular o lucro simulado: %w", err)
	}

	valorFinal := (custoMedio + (pi.CustoMedioNF * 0.01) + (cf.Frete - (cf.Frete * pm.Rebate))) / imposto
	logrus.WithField("valorFinal", valorFinal).Info("Valor final calculado")

	return valorFinal, simulatorProfit, nil
}