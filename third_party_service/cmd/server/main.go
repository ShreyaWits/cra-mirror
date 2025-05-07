package main


func main(){

	 // Load config
	 modules.RegisterModules()

	 kafkaBrokers := strings.Split(os.Getenv("KAFKA_BROKERS"), ",")
	 go thirdpartyservices.StartKafkaConsumer(kafkaBrokers)
 
	 select {} // keep process alive

}