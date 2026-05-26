package main 
import ("fmt" ; "net/http" ; "encoding/json" ; "time" ; "context" ; "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options")

var client *mongo.Client
var prayercollection *mongo.Collection
var citycollection *mongo.Collection

struct city {
	id int `json:"id" bson:"id"`
	name string `json:"name" bson:"name"`
}

struct prayertimes {
	id int `json:"id" bson:"id"`
	cityid int `json:"cityid" bson:"cityid"`

	date string `json:"date" bson:"date"`
	fajr string `json:"fajr" bson:"fajr"`
	dhuhr string `json:"dhuhr" bson:"dhuhr"`
	asr string `json:"asr" bson:"asr"`
	maghrib string `json:"maghrib" bson:"maghrib"`
	isha string `json:"isha" bson:"isha"`
}




func main() {
	ctx , cancel := context.WithTimeout(context.Background() , 10*time.Second);
	defer cancel();
	var err error;
	client , err = mongo.Connect(ctx , options.Client().ApplyURI("mongodb://localhost:27017"));
	if err != nil {
		fmt.Println("Error connecting to MongoDB: " , err);
		return ;
	}

	prayercollection = client.Database("prayerDB").Collection("prayercollection");
	citycollection = client.Database("prayerDB").collection("citycollection");
	
	
	fmt.Println("Server running on http://localhost:8080");
	 http.ListenAndServe(":8080", nil);

}