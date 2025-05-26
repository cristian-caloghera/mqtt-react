package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Brokers []BrokerConfig `yaml:"brokers"`
}

type BrokerConfig struct {
	Address string        `yaml:"address"`
	Topics  []TopicConfig `yaml:"topics"`
}

type TopicConfig struct {
	Name    string `yaml:"name"`
	Command string `yaml:"command"`
}

func (tc *TopicConfig) MessageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message: %s\n", msg.Topic())
	// fire and forget command
	go executeCommand(tc.Command, msg.Payload())
}

// this has some weakness
//   - executing inside a shell to not deal with arguments parsing and white-space escape
func executeCommand(command string, payload []byte) {
	ps := string(payload[:])
	payloadedCommand := strings.ReplaceAll(command, "$payload", ps)
	cmd := exec.Command("sh", "-c", payloadedCommand)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Error executing command '%s': %v\n%s", command, err, string(output))
		return
	}
	log.Printf("%s", string(output))
}

func connectAndSubscribe(brokerConfig BrokerConfig, wg *sync.WaitGroup) {
	defer wg.Done()

	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerConfig.Address)
	opts.SetOnConnectHandler(func(client mqtt.Client) {
		log.Printf("Connected to: %s\n", brokerConfig.Address)
		for i := range brokerConfig.Topics {
			topic := &brokerConfig.Topics[i]
			// MQTT QoS values:
			//   0: at most once ->  offers "fire and forget" messaging with no acknowledgment from the receiver.
			//   1: at least once -> ensures that messages are delivered at least once by requiring a PUBACK acknowledgment.
			//   2: exactly once -> guarantees that each message is delivered exactly once by using a four-step handshake
			var qos byte = 1
			token := client.Subscribe(topic.Name, qos, topic.MessageHandler)
			token.Wait()
			if token.Error() != nil {
				log.Printf("Error subscribing to topic '%s' on broker '%s': %v\n", topic.Name, brokerConfig.Address, token.Error())
				return
			}
			log.Printf("Subscribing to: %s/%s", brokerConfig.Address, topic.Name)
		}
	})
	opts.SetConnectionLostHandler(func(client mqtt.Client, err error) {
		log.Printf("Connection lost from MQTT broker '%s': %v\n", brokerConfig.Address, err)
	})

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Printf("Error connecting to MQTT broker '%s': %v\n", brokerConfig.Address, token.Error())
		return
	}

	// Keep the connection alive
	select {}
}

func main() {
	configFilePath := "./mqtt-react.yaml"

	if len(os.Args) > 1 {
		configFilePath = os.Args[1]
	}
	log.Printf("Using config file: %s", configFilePath)

	// Read configuration from YAML file
	yamlFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatalf("Error reading YAML file: %v\n", err)
	}

	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Error unmarshalling YAML: %v\n", err)
	}

	var wg sync.WaitGroup
	for _, broker := range config.Brokers {
		wg.Add(1)
		go connectAndSubscribe(broker, &wg)
	}

	log.Println("Execution start")
	wg.Wait()
	log.Println("All brokers disconnected. Exiting.")
}
