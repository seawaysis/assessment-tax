package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

type income struct {
	TotalIncome float32     `json:"totalIncome" validate:"required"`
	Wht         float32     `json:"wht"`
	Allowances  []allowance `json:"allowances"`
}
type allowance struct {
	AllowanceType string  `json:"allowanceType"`
	Amount        float32 `json:"amount"`
}
type tax struct {
	Tax      float32    `json:"tax"`
	TaxLevel []taxLevel `json:"taxLevel"`
}
type taxLevel struct {
	Level string  `json:"level"`
	Tax   float32 `json:"tax"`
}
type taxLevelPass []struct {
	Level string  `json:"level"`
	Tax   float32 `json:"tax"`
}
type reciveAmount struct {
	Amount float32 `json:"Amount"`
}
type personalDe struct {
	PersonalDeduction float32 `json:"personalDeduction"`
}
type jwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.RegisteredClaims
}
type incomeMultiple []struct {
	TotalIncome float32
	Wht         float32
	Allowances  []allowance
}
type taxesCSV struct {
	Taxes []taxes
}
type taxes struct {
	TotalIncome float32 `json:"totalIncome"`
	Tax         float32 `json:"tax"`
}

// type reciveDeductionsDB struct {
// 	Category []string
// 	Amount   []float32
// }

func main() {
	e := echo.New()
	e.Use()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})
	e.POST("/tax/cal", cal, somemiddleware)
	e.POST("/tax/calculations", calculations, somemiddleware)
	e.POST("tax/calculations/upload-csv", uploadDeducateFile, somemiddleware)
	e.POST("/admin/deductions/personal", updateDeducatePerson, AuthAdmin)
	go func() {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err)
		}
		prepareDB()
		e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", os.Getenv("PORT"))))
	}()
	//graceful shutdown
	gracefulStop := make(chan os.Signal, 1)
	signal.Notify(gracefulStop, os.Interrupt, syscall.SIGTERM)
	//signal.Notify(gracefulStop, os.Interrupt, syscall.SIGINT)

	<-gracefulStop
	fmt.Println("Server shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		fmt.Printf("Error shutting down server %s", err)
	} else {
		fmt.Println("shut down the server")
	}

}
func somemiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		//fmt.Println("SomeMiddleware")
		/*return c.JSON(
			http.StatusBadGateway,
			map[string]any{"message": "error"},
		)*/ // if error
		return next(c)
	}
}
func AuthAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authorizationToken := c.Request().Header.Get("Authorization")
		if authorizationToken == "" {
			// user := c.FormValue("username")
			// pass := c.FormValue("password")
			user := os.Getenv("ADMIN_USERNAME")
			pass := os.Getenv("ADMIN_PASSWORD")
			if user != "adminTax" || pass != "admin!" {
				return c.JSON(http.StatusBadGateway, "username or password is wrong !!")
			} else {
				return c.JSON(http.StatusUnauthorized, genTokenLogin())
			}
		} else {
			statusCode, err := checkTokenAdmin(authorizationToken)
			if statusCode != 200 {
				return c.JSON(http.StatusUnauthorized, fmt.Sprintf("err => %v | %s", err, genTokenLogin()))
			}
			c.Set("User", err)
			return next(c)
		}
	}
}
func checkTokenAdmin(author string) (int, any) {
	parts := strings.Split(author, " ")
	if !(len(parts) == 2 && parts[0] == "Bearer") {

		return http.StatusUnauthorized, "Token Invalid"
	}
	jwtToken := parts[1]

	token, err := jwt.ParseWithClaims(jwtToken, &jwtCustomClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("secret"), nil
	})
	if err != nil {
		return http.StatusUnauthorized, err
	}
	_, ok := token.Claims.(*jwtCustomClaims)
	if !ok {
		return http.StatusUnauthorized, ok
	}
	return http.StatusOK, token
}
func genTokenLogin() any {
	claims := &jwtCustomClaims{
		"admin",
		true,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Second * 300)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte("secret"))
	if err != nil {
		return err
	}
	return fmt.Sprintf("Bearer token => %v", t)
}
func cal(c echo.Context) error {
	return c.JSON(http.StatusOK, "ok call")
}
func calculations(c echo.Context) error {
	inc := new(income)
	err := c.Bind(&inc)
	if err != nil {
		return c.JSON(http.StatusBadGateway, err)
	} else if inc.TotalIncome < 0 || inc.Wht < 0 {
		return c.JSON(http.StatusBadGateway, "Please add positive number")
	}
	incomeData := struct {
		TotalIncome float32
		Wht         float32
		Allowances  []allowance
	}{
		TotalIncome: inc.TotalIncome,
		Wht:         inc.Wht,
		Allowances:  []allowance{allowance{AllowanceType: inc.Allowances[0].AllowanceType, Amount: inc.Allowances[0].Amount}},
	}
	var incomeSlice incomeMultiple
	incomeSlice = append(incomeSlice, incomeData)

	r := caltaxes(incomeSlice, "level")
	return c.JSON(http.StatusOK, r)
}
func caltaxes(incomeSlice incomeMultiple, types string) any {
	switch types {
	case "level":
		t := new(tax)
		t = &tax{Tax: 0.0}
		for _, m := range incomeSlice {
			numTax := calDeduction(m.TotalIncome, m.Allowances)

			sum, l := calTaxStep(numTax)
			sum = calWht(m.Wht, sum)
			for _, v := range l {
				//parts := strings.Split(v, " ")
				t.Tax = sum
				t.TaxLevel = append(t.TaxLevel, taxLevel{Level: v.Level, Tax: v.Tax})
			}
		}
		return t
	case "CSV":
		t := new(taxesCSV)
		for _, m := range incomeSlice {
			numTax := calDeduction(m.TotalIncome, m.Allowances)

			sum, _ := calTaxStep(numTax)
			sum = calWht(m.Wht, sum)
			t.Taxes = append(t.Taxes, taxes{TotalIncome: m.TotalIncome, Tax: sum})
		}
		return t
	default:
		t := new(tax)
		return t
	}
}
func calDeduction(total float32, m []allowance) float32 {
	// how to cal
	// 500,000 (รายรับ) -
	// 60,0000 (ค่าลดหย่อนส่วนตัว) -
	// 100,000 (เงินบริจาค) -
	// 50,000 (k-receipt)
	// = 290,000
	numTax := total - 60000.0
	for _, v := range m {
		//fmt.Printf("type => %s | amount => %.1f", v.AllowanceType, v.Amount)
		switch v.AllowanceType {
		case "donation":
			if v.Amount > 100000.0 {
				numTax = numTax - 100000.0
			} else {
				numTax = numTax - v.Amount
			}
		case "k-receipt":
			numTax = numTax - v.Amount
		default:
		}
	}
	return numTax
}
func calTaxStep(numTax float32) (float32, taxLevelPass) {
	l := []struct {
		textLevel string
		min       float32
		max       float32
		rate      int
	}{
		{textLevel: "0-150,000", min: 1, max: 150000, rate: 0},
		{textLevel: "150,001-500,000", min: 150001, max: 500000, rate: 10},
		{textLevel: "500,001-1,000,000", min: 500001, max: 1000000, rate: 15},
		{textLevel: "1,000,001-2,000,000", min: 1000001, max: 2000000, rate: 20},
		{textLevel: "2,000,001 ขึ้นไป", min: 2000001, max: 10000000000, rate: 35},
	}
	var sumTax, tempTax, temp float32
	var t taxLevelPass
	for _, v := range l {
		if numTax >= v.min {
			if numTax > v.max {
				temp = v.max - (v.min - 1)
			} else {
				temp = (numTax - (v.min - 1))
			}
			tempTax = (temp * float32(v.rate) / 100)
			sumTax = sumTax + tempTax
			//fmt.Printf("%.1f | %.1f | %.1f\n", numTax, temp, t.Tax)
		} else {
			tempTax = 0
		}
		t = append(t, taxLevel{Level: v.textLevel, Tax: tempTax})
		//t.TaxLevel = append(t.TaxLevel, taxLevel{Level: v.textLevel, Tax: tempTax})
	}
	return sumTax, t
}
func calWht(wht float32, tax float32) float32 {
	if wht > 0 {
		tax = tax - wht
		if tax < 0 {
			tax = 0
		}
	}
	return tax
}
func updateDeducatePerson(c echo.Context) error {
	re := new(reciveAmount)
	err := c.Bind(&re)
	if err != nil {
		return c.JSON(http.StatusBadGateway, err)
	}
	if re.Amount < 10000 || re.Amount > 100000 {
		return c.JSON(http.StatusBadGateway, "personaldeducation must over 10,000 or less 100,000")
	}
	updateDeductions(re.Amount, "personal")
	return c.JSON(http.StatusOK, &personalDe{PersonalDeduction: re.Amount})
}
func connDB() *sql.DB {
	connectionStr := fmt.Sprintf("user=%s password=%s dbname=%s port=5432 sslmode=disable", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"))
	conn, err := sql.Open("postgres", connectionStr)
	if err != nil {
		panic(err)
	}
	return conn
}
func prepareDB() {
	conn := connDB()
	_, err := conn.Exec("CREATE TABLE IF NOT EXISTS deductions (id SERIAL PRIMARY KEY, category TEXT, amount DOUBLE PRECISION);")
	if err != nil {
		log.Fatal("can't create table ", err)
	}
	_, err = conn.Exec("TRUNCATE deductions;")
	if err != nil {
		log.Fatal("can't create table ", err)
	}
	_, err = conn.Exec("INSERT INTO deductions (category, amount) values ($1,$2),($3,$4);", "personal", 60000.0, "kReceipt", 50000.0)
	if err != nil {
		log.Fatal("can't insert default data ", err)
	}
	defer conn.Close()
}
func getDataDeduction(ca string) {
	conn := connDB()
	stmt, err := conn.Prepare("SELECT category,amount FROM deductions WHERE category = $1;")
	if err != nil {
		log.Fatal("can't prepare data ", err)
	}
	rows, err := stmt.Query(ca)
	if err != nil {
		log.Fatal("can't query data ", err)
	}
	for rows.Next() {
		var category string
		var amount float32
		rows.Scan(&category, &amount)
		fmt.Println(category, amount)
	}
}
func updateDeductions(num float32, s string) {
	conn := connDB()
	// stmt, err := conn.Prepare("SELECT id,category,amount FROM deductions")
	// if err != nil {
	// 	log.Fatal("can't prepare data ", err)
	// }

	// rows, err := stmt.Query()
	// if err != nil {
	// 	log.Fatal("can't query data ", err)
	// }
	// for rows.Next() {
	// 	var id string
	// 	var category string
	// 	var amount int
	// 	rows.Scan(&id, &category, &amount)
	// 	fmt.Println(id, category, amount)
	//}
	stmt, err := conn.Prepare("UPDATE deductions SET amount=$1 WHERE category = $2")
	if err != nil {
		log.Fatal("can't prepare data ", err)
	}
	_, err = stmt.Exec(num, s)
	if err != nil {
		log.Fatal("can't update data ", err)
	}
}
func uploadDeducateFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		return err
	}
	inc := setCSV(records)
	t := caltaxes(inc, "CSV")
	return c.JSON(http.StatusOK, t)
}
func setCSV(records [][]string) incomeMultiple {
	var incomeSlice incomeMultiple
	nameCol := [][]string{{"totalIncome", "wht", "donation"}, {"totalIncome", "wht", "k-receipt"}, {"totalIncome", "wht", "donation", "k-receipt"}, {"totalIncome", "wht", "k-receipt", "donation"}}
	checkSchema := 4
	for i, col := range records {
		if i == 0 {
			for j, val := range nameCol {
				if reflect.DeepEqual(col, val) {
					checkSchema = j
					break
				}
			}
		} else {
			if checkSchema != 4 {
				lenSlice := len(nameCol[checkSchema])
				colFloat := convertToFloat(col, lenSlice)
				switch lenSlice {
				case 3:
					allowance1 := allowance{AllowanceType: nameCol[checkSchema][2], Amount: colFloat[2]}
					incomeData := struct {
						TotalIncome float32
						Wht         float32
						Allowances  []allowance
					}{
						TotalIncome: colFloat[0],
						Wht:         colFloat[1],
						Allowances:  []allowance{allowance1},
					}
					incomeSlice = append(incomeSlice, incomeData)
				case 4:
					allowance1 := allowance{AllowanceType: nameCol[checkSchema][2], Amount: colFloat[2]}
					allowance2 := allowance{AllowanceType: nameCol[checkSchema][3], Amount: colFloat[3]}
					incomeData := struct {
						TotalIncome float32
						Wht         float32
						Allowances  []allowance
					}{
						TotalIncome: colFloat[0],
						Wht:         colFloat[1],
						Allowances:  []allowance{allowance1, allowance2},
					}
					incomeSlice = append(incomeSlice, incomeData)
				default:
				}
			}
		}
	}
	return incomeSlice
}
func convertToFloat(str []string, lenSlice int) []float32 {
	newar := make([]float32, lenSlice)
	for i, v := range str {
		value, err := strconv.ParseFloat(v, 32)
		if err != nil {
			fmt.Printf("%v", err)
		}
		newar[i] = float32(value)
	}
	return newar
}
