static const uint8_t analog_pins[] = {A0,A1,A2,A3,A4,A5,A6,A7};
static const uint8_t num_buttons = 8;
static const uint8_t btn_leds[] = {2,3,4,5,6,7,8,9};
static const uint8_t num_leds = 8;
static bool block[] = {false,false,false,false,false,false,false,false};
static unsigned long timers[] = {0,0,0,0,0,0,0,0};
unsigned long blink_time_ms = 200;
int incomingInt; 

void setup() {
  Serial.begin(115200);
  for (int i = 0; i < num_leds; i++){
    pinMode(btn_leds[i], OUTPUT);
  }

  for (int i = 0; i < num_leds; i++){
    digitalWrite(btn_leds[i], HIGH);
    delay(25);
  }
  for (int i = 0; i < num_leds; i++){
    digitalWrite(btn_leds[i], LOW);
    delay(25);
  }
  
  Serial.setTimeout(50);
  int i = 0;
  Serial.write("s");
  //do nice waiting loop
  while (Serial.available() == 0){
    digitalWrite(btn_leds[i%num_leds], HIGH);
    digitalWrite(btn_leds[(i-1)%num_leds], LOW);
    delay(500);
    i++;
  }
  char ch = Serial.read();
  //turn on all lights
  for (int i = 0; i < num_leds; i++) {
    digitalWrite(btn_leds[i], HIGH);
  }
  delay (100);
  //turn off all lights
  for (int i = 0; i < num_leds; i++) {
    digitalWrite(btn_leds[i], LOW);
  }
}

void loop() {
  //check incoming command to leds
  //note that as serial read of 0 can be error leds numbering is 1 based
  if (Serial.available() > 0){
     incomingInt = Serial.parseInt();
     //as 0 is error, buttons numbering starts at 1
     if (1 <= incomingInt <= num_leds){ 
      //toggle led state
      String state = Serial.readString();
      if (state == "H") {
        digitalWrite(btn_leds[incomingInt-1],HIGH);
      } 
      else if (state == "L") {
        digitalWrite(btn_leds[incomingInt-1],LOW);
        timers[incomingInt-1] = 0;
      } 
      else if (state == "B") {
        digitalWrite(btn_leds[incomingInt-1],HIGH);
        timers[incomingInt-1] = millis();
      } 
      else {
        digitalWrite(btn_leds[incomingInt-1],!digitalRead(btn_leds[incomingInt-1]));
      }
    }
  }
  //blink if needed
  for (int i = 0; i < num_leds; i++){
    if (timers[i] == 0) {
      continue;
    }
    if (millis() > timers[i] + blink_time_ms) {
      digitalWrite(btn_leds[i],!digitalRead(btn_leds[i]));
      timers[i] = millis();
      
    }
  }
  //check buttons state
  for (int i = 0; i < num_buttons; i++) {
    if (analogRead(analog_pins[i]) >= 512){
      if (!block[i]){
        //to match between button numbering and leds numbering starts at 1 too
        Serial.write(i+1);
        block[i] = true;
      }
    } else {
      block[i] = false;
    }
    
  }
}
