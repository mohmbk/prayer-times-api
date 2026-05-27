package main 
import ("fmt" ; "net/http" ; "encoding/json" ; "time" ; "context" ; "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options" ; "strconv"
    "strings" )

var client *mongo.Client
var prayercollection *mongo.Collection
var citycollection *mongo.Collection

type city struct {
	id int `json:"id" bson:"id"`
	name string `json:"name" bson:"name"`
}

type prayer struct {
	id int `json:"id" bson:"id"`
	cityid int `json:"cityid" bson:"cityid"`

	date string `json:"date" bson:"date"`
	fajr string `json:"fajr" bson:"fajr"`
	dhuhr string `json:"dhuhr" bson:"dhuhr"`
	asr string `json:"asr" bson:"asr"`
	maghrib string `json:"maghrib" bson:"maghrib"`
	isha string `json:"isha" bson:"isha"`
}


type currentPrayerTimesResponse struct {
	currentPrayer string `json:"current_prayer"`
	nextPrayer string `json:"next_prayer"`
}

func getCurrentDayPrayerTimes(w http.ResponseWriter , r *http.Request) {	

	if r.Method != "GET" {
		http.Error(w , "Method not allowed" , http.StatusMethodNotAllowed)
		return ;
	}

	cityname := r.URL.Query().Get("name");
	var city city ;
	err := citycollection.FindOne(context.Background() , bson.M{"name": cityname}).Decode(&city);
	if err != nil {
		http.Error(w , "City not found" , http.StatusNotFound);
		return ;
	}
	
	now := time.Now();
	today := now.Format("2006-01-02");
	
	var prayer prayer ;
	err = prayercollection.FindOne(context.Background() , bson.M{"cityid": city.id , "date": today}).Decode(&prayer);
	if err != nil {
		http.Error(w , "Prayer times not found for today" , http.StatusNotFound)
		return ;
	}
	

	w.Header().Set("Content-Type" , "application/json");
	json.NewEncoder(w).Encode(prayer);
	
}

func toMinutes(t string) int {
    parts := strings.Split(t, ":");
    hours, _ := strconv.Atoi(parts[0]);
    mins, _  := strconv.Atoi(parts[1]);

    return hours*60 + mins ;
}

func getCurrentPrayerTimes(w http.ResponseWriter , r *http.Request) {
	if r.Method != "GET" {
		http.Error(w , "methodnot allowed" , http.StatusMethodNotAllowed);
		return ;
	}

	now := time.Now();
	currentMinutes := now.Hour()*60 + now.Minute();
	cityname := r.URL.Query().Get("name");

	var city city ;
	err := citycollection.FindOne(context.Background() , bson.M{"name": cityname}).Decode(&city);
	if err != nil {
		http.Error(w , "City not found" , http.StatusNotFound);
		return ;
	}
	today := now.Format("2006-01-02");
	
	var prayer prayer ;
	err = prayercollection.FindOne(context.Background() , bson.M{"cityid": city.id , "date": today}).Decode(&prayer);
	if err != nil {
		http.Error(w , "Prayer times not found for today" , http.StatusNotFound)
		return ;
	}

	fajrMinutes := toMinutes(prayer.fajr);
	dhuhrMinutes := toMinutes(prayer.dhuhr);
	asrMinutes := toMinutes(prayer.asr);
	maghribMinutes := toMinutes(prayer.maghrib);
	ishaMinutes := toMinutes(prayer.isha);

	var currentprayer string ;
	var nextprayer string ;

	if currentMinutes < fajrMinutes {
		currentprayer = "Isha";
		nextprayer = "Fajr";
	} else if currentMinutes < dhuhrMinutes {
		currentprayer = "Fajr";
		nextprayer = "Dhuhr";
	} else if currentMinutes < asrMinutes {
		currentprayer = "Dhuhr";
		nextprayer = "Asr";
	} else if currentMinutes < maghribMinutes {
		currentprayer = "Asr";
		nextprayer = "Maghrib";
	} else if currentMinutes < ishaMinutes {
		currentprayer = "Maghrib";
		nextprayer = "Isha";
	} else {
		currentprayer = "Isha";
		nextprayer = "Fajr";
	}


	var response currentPrayerTimesResponse ;
	response.currentPrayer = currentprayer;
	response.nextPrayer = nextprayer;
	w.Header().Set("Content-Type" , "application/json");
	json.NewEncoder(w).Encode(response);

}


func GetMonth(month string , year string) string {
    return year + "-" + month ;
}


func getprayerTimesForMonth(w http.ResponseWriter , r *http.Request) {
	if r.Method != "GET" {
		http.Error(w , "Method not allowed" , http.StatusMethodNotAllowed);
		return ;
	}
	var cityname string = r.URL.Query().Get("name");
	var month string = r.URL.Query().Get("month");
	var year string = r.URL.Query().Get("year");
	prefix := GetMonth(month , year);
	var city city ;
	err := citycollection.FindOne(context.Background() , bson.M{"name": cityname}).Decode(&city);
	if err != nil {
		http.Error(w , "City not found" , http.StatusNotFound);
		return ;
	}

	cursor , err := prayercollection.Find(context.Background() , bson.M{"cityid": city.id , "date": bson.M{"$regex": "^" + prefix}});
	if err != nil {
		http.Error(w , "Error fetching prayer times" , http.StatusInternalServerError);
		return ;
	}
	defer cursor.Close(context.Background());
	var prayers []prayer ;
	for cursor.Next(context.Background()) {
		var p prayer ;
		cursor.Decode(&p);
		prayers = append(prayers , p);
	}

	w.Header().Set("Content-Type" , "application/json");
	json.NewEncoder(w).Encode(prayers);
	
}


func createCity(w http.ResponseWriter , r *http.Request) {
	if r.Method != "POST" {
		http.Error(w , "Method not allowed" , http.StatusMethodNotAllowed);
		return ;
	}
	var newCity city ;
	err := json.NewDecoder(r.Body).Decode(&newCity);
	if err != nil {
		http.Error(w , "Invalid request body" , http.StatusBadRequest);
		return ;
	}
	result , err := citycollection.InsertOne(context.Background() , newCity);
	if err != nil {
		http.Error(w , "Error creating city" , http.StatusInternalServerError);
		return ;
	}
	fmt.Fprintf(w , "City created with ID: %v" , result.InsertedID);

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
	citycollection = client.Database("prayerDB").Collection("citycollection");

	//for test to see in mongodb compass
	result , err := prayercollection.InsertOne(context.Background() , prayer{id: 1 , cityid: 1 , date: "2024-06-01" , fajr: "04:30" , dhuhr: "12:00" , asr: "15:30" , maghrib: "18:45" , isha: "20:15"});
	fmt.Println("Inserted prayer times with ID: " , result.InsertedID);
	result , err = citycollection.InsertOne(context.Background() , city{id: 1 , name: "Alger"});

	http.HandleFunc("/current_prayer_times" , getCurrentPrayerTimes);
	http.HandleFunc("/today_prayer_times" , getCurrentDayPrayerTimes);
	http.HandleFunc("/month_prayer_times" , getprayerTimesForMonth);
	http.HandleFunc("/create_city" , createCity);

	fmt.Println("Server running on http://localhost:8080");
	 http.ListenAndServe(":8080", nil);

}