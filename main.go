package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	//"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/minio/minio-go"
	"github.com/rs/cors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ItemGet struct {
	ID            primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	Price 		  float64			`json:"price"`
	Images        []string           `bson:"images" json:"images"`
	Type          string             `json:"type"`
	Status        string             `json:"status"`
	
}

type ItemGetInv struct {
	ID            primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	Images        []string           `bson:"images" json:"images"`
	Qty			  int64				 `json:"qty"`
	Price 		  float64			`json:"price"`
	TotalP        float64			`json:"totalp"`
	Type          string             `json:"type"`
	Status        string             `json:"status"`
	
}
type Item struct {
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Price 		  float64			`json:"price"`
	Images        []string          `bson:"images" json:"images"`
	Type          string            `json:"type"`
	Status        string            `json:"status"`
	
}

type PaymentCaptureGet struct {
	ID            primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	CustomerID  string             `json:"custId"`
	Amount      float64            `json:"amount"`
	StripeID    string				`json:"stripeid"`
	Mode        string             `json:"mode"`
	CapturedTimestamp time.Time `bson:"timestamp"`
}

type PaymentCapture struct {
	CustomerID  string             `json:"custId"`
	Amount      float64            `json:"amount"`
	StripeID    string				`json:"stripeid"`
	Mode        string             `json:"mode"`
	CapturedTimestamp time.Time `bson:"timestamp"`
}
type Customer struct {
	Name          string            `json:"name"`
	Careof        string            `json:"careof"`
	Address 	  string            `json:"address"`
	Balance       float64           `json:"balance"`
	Description   string			`json:"description"`
	Number        float64			`json:"number"`
	MonthlypayF   float64          	`json:"monthlypayf"`
	MonthlypayR   float64          	`json:"monthlypayr"`
	DueDay 		  int64				`json:"dueday"`
	Status        string            `json:"status"`
	CapturedTimestamp time.Time `bson:"timestamp"`
}

type CustomerGet struct {
	ID            primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Name          string            `json:"name"`
	Careof        string            `json:"careof"`
	Address 	  string            `json:"address"`
	Balance       float64           `json:"balance"`
	Description   string			`json:"description"`
	Number        float64			`json:"number"`
	MonthlypayF   float64          	`json:"monthlypayf"`
	MonthlypayR   float64          	`json:"monthlypayr"`
	DueDay 		  int64				`json:"dueday"`
	Status        string            `json:"status"`
	CapturedTimestamp time.Time `bson:"timestamp"`
}

type Invoice struct {
	Status        string            `json:"status"`
	Date          time.Time         `bson:"timestamp"`
	Customer      CustomerGet       `json:"customer"`
	Items         []ItemGetInv		`json:"items"`
	Total         float64           `json:"total"`

}

type InvoiceGet struct {
	ID            primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Status        string            `json:"status"`
	Date          time.Time         `bson:"timestamp"`
	Customer      CustomerGet       `json:"customer"`
	Items         []ItemGetInv		`json:"items"`
	Total         float64           `json:"total"`
}


const Database = "omer"


func getEnv(Environment string) (string, error) {
	variable := os.Getenv(Environment)
	if variable == "" {
		fmt.Println(Environment + ` Environment variable is not set`)
		return "", errors.New("env Not Set Properly")
	} else {
		fmt.Printf(Environment+" variable value is: %s\n", variable)
		return variable, nil
	}
}
func main() {
	// MongoDB client options

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},                            // All origins
		AllowedMethods:   []string{"POST", "GET", "PUT", "DELETE"}, // Allowing only get, just an example
		AllowedHeaders:   []string{"Set-Cookie", "Content-Type"},
		ExposedHeaders:   []string{"Set-Cookie"},
		AllowCredentials: true,
		Debug:            true,
	})

	// Get the value of the "ENV_VAR_NAME" environment variable
	mongoURL, err := getEnv("MONGO_URL")
	if err != nil {
		os.Exit(1)
	}

	minioKey, err := getEnv("MINIO_KEY")
	if err != nil {
		os.Exit(1)
	}
	minioSecret, err := getEnv("MINIO_SECRET")
	if err != nil {
		os.Exit(1)
	}

	minioURL, err := getEnv("MINIO_URL")
	if err != nil {
		os.Exit(1)
	}

	

	clientOptions := options.Client().ApplyURI(mongoURL)

	// Create a new MongoDB client
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	minioClient, err := minio.New(minioURL, minioKey, minioSecret, true)
	if err != nil {
		log.Fatalln(err)
	}

	// Create a new router using Gorilla Mux
	router := mux.NewRouter()

	// Define a POST route to add an item to a collection
	router.HandleFunc("/items", addItem(client)).Methods("POST")
	router.HandleFunc("/payment/capture", addPayment(client)).Methods("POST")
	router.HandleFunc("/payment/capture/{id}", addPaymentInvoice(client)).Methods("POST")
	router.HandleFunc("/invoices", addInvoice(client)).Methods("POST")
	router.HandleFunc("/payments", getPayments(client)).Methods("GET")
	router.HandleFunc("/items", getItems(client)).Methods("GET")
	router.HandleFunc("/invoices", getInvoices(client)).Methods("GET")

	router.HandleFunc("/items/disabled", getDisabledItems(client)).Methods("GET")

	router.HandleFunc("/items/{id}", getItem(client)).Methods("GET")
	router.HandleFunc("/invoices/{id}", getInvoice(client)).Methods("GET")

	// Define a DELETE route to delete an item from a collection
	router.HandleFunc("/items/{id}", deleteItem(client)).Methods("DELETE")
	router.HandleFunc("/invoices/{id}", deleteInvoice(client)).Methods("DELETE")

	router.HandleFunc("/items/disabled/{id}", disableItem(client)).Methods("DELETE")

	router.HandleFunc("/items/enabled/{id}", enableItem(client)).Methods("GET")
	router.HandleFunc("/invoices/status/{id}/{status}", setInvoiceStatus(client)).Methods("GET")

	router.HandleFunc("/upload", Upload(minioClient, minioURL)).Methods("POST")

	// Define a PUT route to edit an item in a collection
	router.HandleFunc("/items/{id}", editItem(client)).Methods("PUT")
	router.HandleFunc("/payments/{id}", editPayment(client)).Methods("PUT")
	router.HandleFunc("/payments/revert/{id}", revertPayment(client)).Methods("DELETE")
	router.HandleFunc("/invoices/{id}", editInvoice(client)).Methods("PUT")
	

	// Define a POST route to add an item to a collection
	router.HandleFunc("/customer", addCustomer(client)).Methods("POST")

	router.HandleFunc("/customers", getCustomers(client)).Methods("GET")

	router.HandleFunc("/customer/{id}", getCustomer(client)).Methods("GET")

	// Define a DELETE route to delete an item from a collection
	router.HandleFunc("/customer/{id}", deleteCustomer(client)).Methods("DELETE")

	router.HandleFunc("/customer/disabled/{id}", disableCustomer(client)).Methods("DELETE")

	router.HandleFunc("/customer/enabled/{id}", enableCustomer(client)).Methods("GET")

	// Define a PUT route to edit an item in a collection
	router.HandleFunc("/customer/{id}", editCustomer(client)).Methods("PUT")
	
	// Start the HTTP server
	log.Println("Starting HTTP server...")
	err = http.ListenAndServe(":8002", c.Handler(router))
	if err != nil {

		log.Fatal(err)
	}
}


func Upload(minioClient *minio.Client, minioURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the multipart form.
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			fmt.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get the file headers from the form.
		files := r.MultipartForm.File["files"]

		var ImagePaths []string

		// Loop through the files and upload them to Minio.
		for _, fileHeader := range files {
			// Open the file.
			file, err := fileHeader.Open()
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer file.Close()

			// Get the file name and extension.
			filename := fileHeader.Filename
			extension := filepath.Ext(filename)

			dotRemoved := extension[1:]

			// Generate a unique file name with the original extension.
			newFilename := fmt.Sprintf("%d%s", time.Now().UnixNano(), extension)
			newPath := "https://" + minioURL + "/omer/" + newFilename
			ImagePaths = append(ImagePaths, newPath)

			log.Println(ImagePaths)

			// Upload the file to Minio.
			_, err = minioClient.PutObject("omer", newFilename, file, fileHeader.Size, minio.PutObjectOptions{
				ContentType: "image/" + dotRemoved,
			})
			if err != nil {
				fmt.Println(err.Error())
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		data, err := json.Marshal(ImagePaths)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Set the Content-Type header to application/json
		//w.Header().Set("Content-Type", "application/json")
		// Write the JSON data to the response

		// Send a success response.
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

// addItem inserts a new item into the "items" collection in MongoDB
func addItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body into an Item struct
		var item Item
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Println(item)

		// Insert the item into the "items" collection in MongoDB
		collection := client.Database(Database).Collection("products")
		_, err = collection.InsertOne(context.Background(), item)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusCreated)
	}
}

// addItem inserts a new item into the "items" collection in MongoDB
func addInvoice(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body into an Item struct
		var item Invoice
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		item.Date = time.Now()

		log.Println(item)

		// Insert the item into the "items" collection in MongoDB
		collection := client.Database(Database).Collection("invoices")
		_, err = collection.InsertOne(context.Background(), item)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection = client.Database(Database).Collection("customer")
		filter := bson.M{"_id": item.Customer.ID}
		previousCust := CustomerGet{}
		err = collection.FindOne(context.Background(), bson.M{"_id": item.Customer.ID}).Decode(&previousCust)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newBalance := previousCust.Balance + item.Total
		update := bson.M{"$set": bson.M{"balance": newBalance}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusCreated)
	}
}

func addPaymentInvoice(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection := client.Database(Database).Collection("invoices")
		filter := bson.M{"_id": oid}
		update := bson.M{"$set": bson.M{"status": "paid"}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Parse the request body into an Item struct
		var item PaymentCapture
		err = json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		item.CapturedTimestamp = time.Now()

		log.Println(item)

		// Insert the item into the "items" collection in MongoDB
		collection = client.Database(Database).Collection("payments")
		_, err = collection.InsertOne(context.Background(), item)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		oid, err = primitive.ObjectIDFromHex(item.CustomerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection = client.Database(Database).Collection("customer")
		filter = bson.M{"_id": oid}
		previousCust := CustomerGet{}
		err = collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&previousCust)
		if err != nil {                                                                                        
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newBalance := previousCust.Balance - item.Amount
		if newBalance <= 0 {

		}
		update = bson.M{"$set": bson.M{"balance": newBalance}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusCreated)
	}
}


// addItem inserts a new item into the "items" collection in MongoDB
func addPayment(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body into an Item struct
		var item PaymentCapture
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		item.CapturedTimestamp = time.Now()

		log.Println(item)

		// Insert the item into the "items" collection in MongoDB
		collection := client.Database(Database).Collection("payments")
		_, err = collection.InsertOne(context.Background(), item)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		oid, err := primitive.ObjectIDFromHex(item.CustomerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection = client.Database(Database).Collection("customer")
		filter := bson.M{"_id": oid}
		previousCust := CustomerGet{}
		err = collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&previousCust)
		if err != nil {                                                                                        
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		newBalance := previousCust.Balance - item.Amount
		if newBalance <= 0 {

		}
		update := bson.M{"$set": bson.M{"balance": newBalance}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusCreated)
	}
}

// addItem inserts a new item into the "items" collection in MongoDB
func addCustomer(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse the request body into an Item struct
		var item Customer
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Println(item)
		item.CapturedTimestamp = time.Now()
		log.Println(item)

		// Insert the item into the "items" collection in MongoDB
		collection := client.Database(Database).Collection("customer")
		_, err = collection.InsertOne(context.Background(), item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusCreated)
	}
}


// deleteItem deletes an item from the "items" collection in MongoDB
func deleteItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		// Delete the item from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("products")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = collection.DeleteOne(context.Background(), bson.M{"_id": oid})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

// deleteItem deletes an item from the "items" collection in MongoDB
func deleteInvoice(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		// Delete the item from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("invoices")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = collection.DeleteOne(context.Background(), bson.M{"_id": oid})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

// deleteItem deletes an item from the "items" collection in MongoDB
func deleteCustomer(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		// Delete the item from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("customer")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = collection.DeleteOne(context.Background(), bson.M{"_id": oid})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

// editItem updates an item in the "items" collection in MongoDB
func editItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		// Parse the request body into an Item struct
		var item ItemGet
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(item)

		// tables := item.TableAttached

		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("products")
		filter := bson.M{"_id": item.ID}
		update := bson.M{"$set": bson.M{"name": item.Name, "description": item.Description, "status": item.Status, "images": item.Images, "type": item.Type, "price": item.Price}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

// editItem updates an item in the "items" collection in MongoDB
func editInvoice(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		// Parse the request body into an Item struct
		var item InvoiceGet
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(item)
		previousInv := InvoiceGet{}
		previousCust := CustomerGet{}
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		collection := client.Database(Database).Collection("invoices")
		err = collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&previousInv)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection = client.Database(Database).Collection("customer")
		err = collection.FindOne(context.Background(), bson.M{"_id": item.Customer.ID}).Decode(&previousCust)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection = client.Database(Database).Collection("customer")
		filter := bson.M{"_id": item.Customer.ID}
		log.Println("cust bal:", previousCust.Balance)
		log.Println("prev inv:", previousInv.Total)
		resetBalance := previousCust.Balance - previousInv.Total
		log.Println("reset:", resetBalance)
		newBalance := resetBalance + item.Total
		log.Println("new bal:", newBalance)
		update := bson.M{"$set": bson.M{"balance": newBalance}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}




		// Update the item in the "items" collection in MongoDB
		collection = client.Database(Database).Collection("invoices")
		filter = bson.M{"_id": item.ID}
		update = bson.M{"$set": bson.M{"total": item.Total, "items": item.Items, "status": item.Status, "customer": item.Customer}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

// editItem updates an item in the "items" collection in MongoDB
func editPayment(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		// Parse the request body into an Item struct
		var item PaymentCaptureGet
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(item)
		previousPayment := PaymentCaptureGet{}
		previousCust := CustomerGet{}
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		collection := client.Database(Database).Collection("payments")
		err = collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&previousPayment)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		coid, err := primitive.ObjectIDFromHex(item.CustomerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection = client.Database(Database).Collection("customer")
		err = collection.FindOne(context.Background(), bson.M{"_id": coid}).Decode(&previousCust)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		
		filter := bson.M{"_id": coid}
		log.Println("cust bal:", previousCust.Balance)
		log.Println("prev amount:", previousPayment.Amount)
		resetBalance := previousCust.Balance + previousPayment.Amount
		log.Println("reset:", resetBalance)
		newBalance := resetBalance - item.Amount
		log.Println("new bal:", newBalance)
		update := bson.M{"$set": bson.M{"balance": newBalance}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}




		// Update the item in the "items" collection in MongoDB
		collection = client.Database(Database).Collection("payments")
		filter = bson.M{"_id": item.ID}
		update = bson.M{"$set": bson.M{"custId": item.CustomerID, "amount": item.Amount, "stripeid": item.StripeID, "mode": item.Mode}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

// editItem updates an item in the "items" collection in MongoDB
func revertPayment(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		previousPayment := PaymentCaptureGet{}
		previousCust := CustomerGet{}
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		collection := client.Database(Database).Collection("payments")
		err = collection.FindOne(context.Background(), bson.M{"_id": oid}).Decode(&previousPayment)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		coid, err := primitive.ObjectIDFromHex(previousPayment.CustomerID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection = client.Database(Database).Collection("customer")
		err = collection.FindOne(context.Background(), bson.M{"_id": coid}).Decode(&previousCust)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		
		filter := bson.M{"_id": coid}
		log.Println("cust bal:", previousCust.Balance)
		log.Println("prev amount:", previousPayment.Amount)
		resetBalance := previousCust.Balance + previousPayment.Amount
		log.Println("reset:", resetBalance)
		update := bson.M{"$set": bson.M{"balance": resetBalance}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		collection = client.Database(Database).Collection("payments")
		_, err = collection.DeleteOne(context.Background(), bson.M{"_id": oid})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

// editItem updates an item in the "items" collection in MongoDB
func editCustomer(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		// Parse the request body into an Item struct
		var item CustomerGet
		err := json.NewDecoder(r.Body).Decode(&item)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println(item)

		// tables := item.TableAttached

		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("customer")
		filter := bson.M{"_id": item.ID}
		update := bson.M{"$set": bson.M{"name": item.Name, "careof": item.Careof, "status": item.Status, "address": item.Address, "number": item.Number, "balance": item.Balance, "dueday": item.DueDay, "monthlypayf": item.MonthlypayF, "monthlypayr": item.MonthlypayR, "description": item.Description}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

func disableItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("products")
		filter := bson.M{"_id": oid}
		update := bson.M{"$set": bson.M{"status": "disabled"}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

func disableCustomer(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("customer")
		filter := bson.M{"_id": oid}
		update := bson.M{"$set": bson.M{"status": "disabled"}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

func enableItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("products")
		filter := bson.M{"_id": oid}
		update := bson.M{"$set": bson.M{"status": "active"}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

func setInvoiceStatus(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]
		status := vars["status"]

		fmt.Println(id)

		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("invoices")
		filter := bson.M{"_id": oid}
		update := bson.M{"$set": bson.M{"status": status}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

func enableCustomer(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the name parameter from the request URL
		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)

		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the item in the "items" collection in MongoDB
		collection := client.Database(Database).Collection("customer")
		filter := bson.M{"_id": oid}
		update := bson.M{"$set": bson.M{"status": "active"}}
		_, err = collection.UpdateOne(context.Background(), filter, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Send a success response
		w.WriteHeader(http.StatusOK)
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getItems(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("products")
		findOptions := options.Find().SetSort(bson.D{{Key: "name", Value: 1}}) 
		cursor, err := collection.Find(context.Background(), bson.M{}, findOptions)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Decode the cursor results into a slice of Item structs
		var items []ItemGet
		for cursor.Next(context.Background()) {
			var item ItemGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getInvoices(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("invoices")
		findOptions := options.Find().SetSort(bson.D{{Key: "name", Value: 1}}) 
		cursor, err := collection.Find(context.Background(), bson.M{}, findOptions)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Decode the cursor results into a slice of Item structs
		var items []InvoiceGet
		for cursor.Next(context.Background()) {
			var item InvoiceGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getPayments(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("payments")
		findOptions := options.Find().SetSort(bson.D{{Key: "_id", Value: 1}}) 
		cursor, err := collection.Find(context.Background(), bson.M{}, findOptions)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Decode the cursor results into a slice of Item structs
		var items []PaymentCaptureGet
		for cursor.Next(context.Background()) {
			var item PaymentCaptureGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getCustomers(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("customer")
		findOptions := options.Find().SetSort(bson.D{{Key: "name", Value: 1}}) 
		cursor, err := collection.Find(context.Background(), bson.M{}, findOptions)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Decode the cursor results into a slice of Item structs
		var items []CustomerGet
		for cursor.Next(context.Background()) {
			var item CustomerGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getDisabledItems(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("products")
		cursor, err := collection.Find(context.Background(), bson.M{"status": "disabled"})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Decode the cursor results into a slice of Item structs
		var items []ItemGet
		for cursor.Next(context.Background()) {
			var item ItemGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getItem(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("products")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cursor, err := collection.Find(context.Background(), bson.M{"_id": oid})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Decode the cursor results into a slice of Item structs
		var items []ItemGet
		for cursor.Next(context.Background()) {
			var item ItemGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getInvoice(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("invoices")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cursor, err := collection.Find(context.Background(), bson.M{"_id": oid})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Decode the cursor results into a slice of Item structs
		var items []InvoiceGet
		for cursor.Next(context.Background()) {
			var item InvoiceGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// getItems retrieves all items from the "items" collection in MongoDB
func getCustomer(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		vars := mux.Vars(r)
		id := vars["id"]

		fmt.Println(id)
		// Get all items from the "items" collection in MongoDB
		collection := client.Database(Database).Collection("customer")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		cursor, err := collection.Find(context.Background(), bson.M{"_id": oid})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cursor.Close(context.Background())

		// Decode the cursor results into a slice of Item structs
		var items []CustomerGet
		for cursor.Next(context.Background()) {
			var item CustomerGet
			err := cursor.Decode(&item)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			items = append(items, item)
		}

		// Send the list of items as a JSON response
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(items)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

