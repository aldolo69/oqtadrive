/*
    OqtaDrive - Sinclair Microdrive emulator
    Copyright (c) 2021, Alexander Vollschwitz

    developed on: Arduino Nano

    This file is part of OqtaDrive.

    OqtaDrive is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    OqtaDrive is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with OqtaDrive. If not, see <http://www.gnu.org/licenses/>.
*/

/*
	Implementation notes:
	- _delay_us only takes compile time constants as argument
 */

// Change this to true for a calibration run. When not connecting the adapter to
// an Interface 1/QL during calibration, choose the desired interface via the
// force settings below.
#define CALIBRATION false

// --- pin assignments --------------------------------------------------------
//
// Note: Changing pin assignments (other than LED pins) will break things!
//       These constants are for clarity & convenience only.
//
const int PIN_COMMS_CLK  = 2; // HIGH idle on IF1, LOW on QL; interrupt
const int PIN_COMMS_IN   = 4;
const int PIN_ERASE      = 5; // LOW active
const int PIN_READ_WRITE = 3; // READ is HIGH; interrupt
const int PIN_WR_PROTECT = 6; // LOW active

const int PIN_LED_WRITE  = 11;
const int PIN_LED_READ   = 12;

const int PIN_TRACK_1 = A4;
const int PIN_TRACK_2 = A0;

// --- pin masks --------------------------------------------------------------
const uint8_t MASK_COMMS_CLK  = 1 << PIN_COMMS_CLK;
const uint8_t MASK_COMMS_IN   = 1 << PIN_COMMS_IN;
const uint8_t MASK_ERASE      = 1 << PIN_ERASE;
const uint8_t MASK_READ_WRITE = 1 << PIN_READ_WRITE;
const uint8_t MASK_WR_PROTECT = 1 << PIN_WR_PROTECT;
const uint8_t MASK_RECORDING  = MASK_ERASE | MASK_READ_WRITE;

const uint8_t MASK_TRACK_1     = B00010000;
const uint8_t MASK_TRACK_2     = B00000001;
const uint8_t MASK_BOTH_TRACKS = MASK_TRACK_1 | MASK_TRACK_2;

const uint8_t MASK_LED_WRITE  = B00001000;
const uint8_t MASK_LED_READ   = B00010000;

// --- LED behavior -----------------------------------------------------------
const bool ACTIVE = true;
const bool IDLE   = false;
// whether read & write LEDs should be on or off during inactivity
const bool LED_RW_IDLE_ON = true;

uint8_t blinkCount = 0;

// --- tape format ------------------------------------------------------------
const int PREAMBLE_LENGTH   = 12;
const int HEADER_LENGTH_IF1 = 27;
const int RECORD_LENGTH_IF1 = 540;
const int HEADER_LENGTH_QL  = 28;
const int RECORD_LENGTH_QL  = 538;
const int RECORD_EXTRA_QL   = 86; // during format, QL sends longer records

uint16_t headerLengthMux;
uint16_t recordLengthMux;
uint16_t sectorLengthMux;

// --- timer pre-loads, based on a 256 pre-scaler, 1 tick is 16us -------------
const int TIMER_COMMS          = 65536 - 625;  //   10 msec
const int TIMER_HEADER_GAP_IF1 = 65536 - 234;  // 3.75 msec
const int TIMER_HEADER_GAP_QL  = 65536 - 225;  // 3.60 msec

// --- drive select -----------------------------------------------------------
volatile uint8_t commsRegister = 0;
volatile uint8_t commsClkCount = 0;
volatile uint8_t activeDrive   = 0;
volatile uint8_t driveOffset   = 0xff;

// Automatic offset check only works for QL. If you're using OqtaDrive with an
// actual Microdrive between IF1 and the adapter, you can set a fixed offset
// here. Likewise for the QL, if the automatic check doesn't work reliably.
// The offset denotes how many actual Microdrives are present between the
// Microdrive interface and the adapter. So an offset of 0 means the adapter is
// directly connected to the IF1 or internal Microdrive interface on the QL,
// bypassing the two built-in drives. Max accepted value is 7. Keep at -1 to use
// automatic offset check.
const int DRIVE_OFFSET_IF1 = 0;
const int DRIVE_OFFSET_QL  = -1;

bool commsClkState;

const uint8_t DRIVE_STATE_UNKNOWN  = 0x80;
const uint8_t DRIVE_FLAG_LOADED    = 1;
const uint8_t DRIVE_FLAG_FORMATTED = 2;
const uint8_t DRIVE_FLAG_READONLY  = 4;
const uint8_t DRIVE_READABLE = DRIVE_FLAG_LOADED | DRIVE_FLAG_FORMATTED;

volatile uint8_t driveState = DRIVE_STATE_UNKNOWN;

// --- interrupt handler ------------------------------------------------------
typedef void (* TimerHandler)();
void setTimer(int preload, TimerHandler h);
void enableTimer(bool on, int preload, TimerHandler h);
TimerHandler timerHandler = NULL;

// --- sector buffer ----------------------------------------------------------
const uint16_t BUF_LENGTH = 10 +
	max(HEADER_LENGTH_IF1, HEADER_LENGTH_QL) +
	max(RECORD_LENGTH_IF1, RECORD_LENGTH_QL + RECORD_EXTRA_QL);
uint8_t buffer[BUF_LENGTH];

// --- message buffer ---------------------------------------------------------
uint8_t msgBuffer[4];

// --- Microdrive interface type - Interface 1 or QL --------------------------
const bool FORCE_IF1 = false;
const bool FORCE_QL  = false;

bool IF1 = true;
#define QL !IF1

// --- state flags ------------------------------------------------------------
volatile bool spinning    = false;
volatile bool recording   = false;
volatile bool message     = false;
volatile bool headerGap   = false;
volatile bool calibration = false; // use the define setting at top to turn on!
volatile bool synced      = false;

// --- daemon commands --------------------------------------------------------
const char CMD_HELLO  = 'h';
const char CMD_PING   = 'P';
const char CMD_STATUS = 's';
const char CMD_GET    = 'g';
const char CMD_PUT    = 'p';
const char CMD_VERIFY = 'v';
const char CMD_DEBUG  = 'd';

const uint8_t  CMD_LENGTH = 4;
const uint16_t PAYLOAD_LENGTH = BUF_LENGTH - CMD_LENGTH;

const char DAEMON_PING[]  = {CMD_PING, 'i', 'n', 'g'};
const char DAEMON_PONG[]  = {CMD_PING, 'o', 'n', 'g'};
const char DAEMON_HELLO[] = {CMD_HELLO, 'l' , 'o', 'd'};
const char IF1_HELLO[]    = {CMD_HELLO, 'l' , 'o', 'i'};
const char QL_HELLO[]     = {CMD_HELLO, 'l' , 'o', 'q'};

const unsigned long DAEMON_TIMEOUT   =  5000;
const unsigned long RESYNC_THRESHOLD =  4500;
const unsigned long PING_INTERVAL    = 10000;

unsigned long lastPing = 0;

// ------------------------------------------------------------------ SETUP ---

void setup() {

	// control signals
	pinMode(PIN_READ_WRITE, INPUT_PULLUP);
	pinMode(PIN_ERASE, INPUT_PULLUP);
	pinMode(PIN_COMMS_CLK, INPUT_PULLUP);
	// This must not be set to INPUT_PULLUP. If there is a Microdrive upstream
	// of the adapter in the daisy chain, the pull-up resistor would feed into
	// that drive's COMMS_CLK output and confuse it.
	pinMode(PIN_COMMS_IN, INPUT);
	//
	pinMode(PIN_WR_PROTECT, OUTPUT);
	digitalWrite(PIN_WR_PROTECT, HIGH);

	// LEDs
	pinMode(PIN_LED_WRITE, OUTPUT);
	pinMode(PIN_LED_READ, OUTPUT);
	ledRead(IDLE);
	ledWrite(IDLE);

	setTracksToRecord();

	// open channel to daemon & say hello
	detectInterface();
	Serial.begin(1000000, SERIAL_8N1); // 1Mbps is highest reliable rate
	Serial.setTimeout(DAEMON_TIMEOUT);

	// set up interrupts
	attachInterrupt(digitalPinToInterrupt(PIN_COMMS_CLK), commsClk,
		IF1 ? FALLING : RISING);
	attachInterrupt(digitalPinToInterrupt(PIN_READ_WRITE), writeReq, FALLING);
}

// ------------------------------------------------------------------- LOOP ---

void loop() {

	// FIXME: verify that serial is only accessed from main loop

	if (CALIBRATION) {
		calibrate(0x0f);
	}

	if (!synced) {
		driveOff();
		daemonSync();
	}

	debugFlush();
	ensureDriveState();

	if (spinning) {

		if (recording) {
			record();
		} else {
			replay();
		}

		if (blinkCount++ % 2 == 0) {
			if (recording) {
				ledRead(IDLE);
				ledWriteFlip();
			} else {
				ledWrite(IDLE);
				ledReadFlip();
			}
		}

		lastPing = millis();

	} else {
		daemonPing();
	}
}

// ---------------------------------------------- Interface 1 / QL HANDLING ---

void detectInterface() {

	if (FORCE_IF1) {
		IF1 = true;

	} else if (FORCE_QL) {
		IF1 = false;

	} else {
		// Idle level of COMMS_CLK is HIGH for Interface 1, LOW for QL.
		// Sample COMMS_CLK line for two seconds to find out.
		uint8_t high = 0, low = 0;
		for (uint8_t i = 0; i < 21; i++) {
			(PIND & MASK_COMMS_CLK) == 0 ? low++ : high++;
			delay(100);
		}
		IF1 = high > low;
	}

	if (IF1) {
		if ((-1 < DRIVE_OFFSET_IF1) && (DRIVE_OFFSET_IF1 < 8)) {
			driveOffset = DRIVE_OFFSET_IF1;
		}
		headerLengthMux = HEADER_LENGTH_IF1 + 1;
		recordLengthMux = RECORD_LENGTH_IF1 + 1;
		commsClkState = true;

	} else {
		if ((-1 < DRIVE_OFFSET_QL) && (DRIVE_OFFSET_QL < 8)) {
			driveOffset = DRIVE_OFFSET_QL;
		}
		headerLengthMux = HEADER_LENGTH_QL + 1;
		recordLengthMux = RECORD_LENGTH_QL + 1;
		commsClkState = false;
	}

	sectorLengthMux = headerLengthMux + recordLengthMux;
}

// ----------------------------------------------------- READ/WRITE CONTROL ---

/*
	Check whether the WRITE or ERASE line is active (indicates recording),
	or an impending drive state change is indicated. Switches recording and
	spinning states accordingly, and returns true if any of the above is the
	case. This is used repeatedly while in replay mode to find out whether we
	need to bail out.

	Note: Any change in this function requires re-calibration of replay!

	TODO: Consider whether to consolidate this with checkReplayOrStop, since
	      most of the code is the same. We may not want to do this though for
	      timing reasons.
 */
bool checkRecordingOrStop() {

	uint8_t state = PIND;

	// turn recording on, but never off here
	recording = recording || ((state & MASK_RECORDING) != MASK_RECORDING);

	// when recording, flip track mode here already
	// to give it more time to switch over
	if (recording) {
		setTracksToRecord();
	}

	// turn spinning off, but never on here; change in drive state is indicated
	// for IF1 by COMMS_CLK going active (i.e. LOW), and for QL by going inactive
	// (also LOW)
	bool prev = commsClkState;
	commsClkState = (state & MASK_COMMS_CLK) != 0;
	spinning = spinning && (!prev || commsClkState);

	return !spinning || recording;
}

/*
	Check whether both WRITE and ERASE lines are inactive (indicates replay),
	or an impending drive state change is indicated. Switches recording and
	spinning states accordingly, and returns true if any of the above is the
	case. This is used repeatedly while in record mode to find out whether we
	need to bail out.
 */
bool checkReplayOrStop() {

	uint8_t state = PIND;

	// turn recording off, but never on here
	recording = recording && ((state & MASK_RECORDING) != MASK_RECORDING);

	if (!recording) {
		setTracksToReplay();
	}

	// turn spinning off, but never on here; change in drive state is indicated
	// for IF1 by COMMS_CLK going active (i.e. LOW), and for QL by going inactive
	// (also LOW)
	bool prev = commsClkState;
	commsClkState = (state & MASK_COMMS_CLK) != 0;
	spinning = spinning && (!prev || commsClkState);

	return !(spinning && recording);
}

// interrupt handler; switches tracks to record mode
void writeReq() {
	if (activeDrive > 0) {
		setTracksToRecord();
	}
}

// interrupt handler; signals the end of the header gap
void endHeaderGap() {
	headerGap = false;
}

/*
	Tracks in RECORD mode means the Arduino reads incoming data. I.e. when the
	Interface 1/QL wants to write to the Microdrive, the two track data pins
	need to be put into input mode.
 */
void setTracksToRecord() {
	DDRC = 0;
	PORTC = 0x3f;
}

/*
	Tracks in REPLAY mode means the Arduino sends data. I.e. when the
	Interface 1/QL wants to read from the Microdrive, the two track data pins
	need to be put into	output mode.
 */
void setTracksToReplay() {
	DDRC = MASK_BOTH_TRACKS;
	PORTC = 0x3f; // idle level is HIGH
}

// ---------------------------------------------------------- DRIVE CONTROL ---

void driveOff() {
	stopTimer();
	setTracksToRecord();
	ledWrite(IDLE);
	ledRead(IDLE);
	spinning = false;
	recording = false;
	headerGap = false;
}

void driveOn() {
	setTracksToReplay();
	headerGap = false;
	driveState = DRIVE_STATE_UNKNOWN;
	recording = false;
	spinning = true;
}

// Active level of COMMS_CLK is LOW for Interface 1, HIGH for QL.
bool isCommsClk(uint8_t state) {
	return (state & MASK_COMMS_CLK) == (IF1 ? 0 : MASK_COMMS_CLK);
}

/*
	Interrupt handler for handling the 1 bit being pushed through the
	Microdrive daisy chain to select the active drive.

	We need to add the next COMMS_IN bit on COMMS_CLK going active, i.e.
	when logically rising. I assume that the h/w is doing this on falling
	clock, but if we were to do this here with interrupt mechanism, we'd
	be too late and	by the time we sample COMMS_IN, it already holds the
	next bit.
 */
void commsClk() {

	if (driveOffset == 0xff && (PIND & MASK_COMMS_IN) != 0) {
		// When we see the 1 bit at COMMS_CLK going active, clock count is the
		// drive offset, with an offset of 0 meaning first drive in chain.
		driveOffset = IF1 ? DRIVE_OFFSET_IF1 : commsClkCount;
	}

	stopTimer();
	commsClkCount++;
	commsRegister = commsRegister << 1;
	commsRegister |= ((PIND & MASK_COMMS_IN) != 0 ? 1 : 0);
	setTimer(TIMER_COMMS, selectDrive);
}

/*
	Called by timer when TIMER_COMMS expires. That is, if for a duration of
	TIMER_COMMS, there has been no change on the COMMS_CLK line, which indicates
	that the active drive has been selected, or all drives deselected, by
	Interface 1/QL.
 */
void selectDrive() {

	if (QL && isCommsClk(PIND)) {
		// QL switches COMMS_CLK to HIGH and keeps it HIGH as long as its
		// interested in reading more data. Should the timer fire when we're
		// still reading, we just discard it.
		return;
	}

	activeDrive = 0;

	// A drive offset of 0xff means we don't know yet, and that also means
	// none of the virtual drives can possibly have been selected.
	if (driveOffset != 0xff) {
		for (uint8_t reg = commsRegister << driveOffset; reg > 0; reg >>= 1) {
			activeDrive++;
		}
	}

	debugMsg('C', 'K', commsClkCount);
	debugFlush();
//	debugMsg('O', 'F', driveOffset);
//	debugFlush();
	debugMsg('R', 'l', lowByte(commsRegister));
	debugFlush();
	debugMsg('R', 'h', highByte(commsRegister));
	debugFlush();
	debugMsg('D', 'R', activeDrive);
	debugFlush();

	if (activeDrive == 0) {
		driveOff();
	} else {
		driveOn();
	}
}

/*
	Retrieve the drive state from the daemon, if it is still unknown. Otherwise,
	cached value is returned.
 */
void ensureDriveState() {

	if (spinning && (driveState != DRIVE_STATE_UNKNOWN)) {
		return;
	}

	if (!spinning && (driveState == DRIVE_STATE_UNKNOWN)) {
		return;
	}

	daemonCmdArgs(CMD_STATUS, activeDrive, spinning ? 1 : 0, 0, 0);

	if (spinning) {
		for (int r = 0; r < 400; r++) {
			if (Serial.available() > 0) {
				driveState = Serial.read();
				return;
			}
			delay(5);
		}
		synced = false;

	} else {
		driveState = DRIVE_STATE_UNKNOWN;
	}
}

//
bool isDriveReadable() {
	return (driveState & DRIVE_READABLE) == DRIVE_READABLE;
}

//
bool isDriveWritable() {
	return isDriveReadable() && ((driveState & DRIVE_FLAG_READONLY) == 0);
}

// -------------------------------------------------------------- RECORDING ---

void record() {

	bool formatting = false;
	uint8_t blocks = 0;
	ledWrite(ACTIVE);

	do {
		daemonPendingCmd(CMD_PUT, activeDrive, 0);
		uint16_t read = receiveBlock();
		blocks++;

		if (blocks % 4 == 0) {
			ledWriteFlip();
		}

 		// block stop marker; used by the daemon to get rid of spurious
 		// extra bytes at the end of a block, often seen on the QL
		if (read > 0) {
			daemonCmdArgs(3, 2, 1, 0, 0);
		}

		read += PREAMBLE_LENGTH; // preamble is not sent to daemon

		if (read < headerLengthMux) {
			break; // nothing useful received
		} else if (read < recordLengthMux) {
			// headers are only written during format
			// when that happens, we stay here
			formatting = true;
			ledRead(IDLE);
		}

	} while (formatting);

	if (formatting) {
		driveState = DRIVE_FLAG_LOADED | DRIVE_FLAG_FORMATTED;
	}

	ledWrite(IDLE);
	checkReplayOrStop();
}

/*
	Receive a block (header or record). Change in drive state is checked only
	while waiting for the start of the block. Once the block starts, no further
	checks are done. End of data from Interface 1/QL however is detected.

	We're reading both tracks simultaneously. Pin assignments are chosen such
	that one track is at bit position 0, the other at 4. 4 bits from each track
	are ORed into `d`, with a left shift before each OR.

                          --------------- track 1
                          |           --- track 2
           bit            |           |
      position:  7  6  5  4  3  2  1  0
          PINC:  X  X  X  |  X  X  X  |   X = don't care
                   A N D  |           |
          MASK:  0  0  0  1  0  0  0  1
                    O R   |           |
             d: [ track 1 *][ track 2 *]  << before each OR

	`d` is then forwarded to the daemon over the serial line. This is repeated
	until the block (header or record) is done. The number of bytes read is
	returned. The receiving side takes care of demuxing the data, additionally
	considering that the tracks are shifted by 4 bits relative to one another.
 */
uint16_t receiveBlock() {

 	noInterrupts();

	register uint8_t start = PINC & MASK_BOTH_TRACKS, end, bitCount, d, w;
	register uint16_t read = 0, ww;

	for (ww = 0xffff; ww > 0; ww--) {
		// sync on first bit change of block, on either track
		if (((PINC & MASK_BOTH_TRACKS) ^ start) != 0) {
			break;
		}
		// but don't wait forever, and bail out when activity on COMMS_CLK
		// indicates impending change of drive state
		if (checkReplayOrStop() || ww == 1) {
			UDR0 = ww == 1 ? 2 : 1; // cancel pending PUT command
			interrupts();
			return 0;
		}
	}

	// search for sync pattern, which is 10*'0' followed by 2*'ff'; we therefore
	// look for at least 24 consecutive zeros on both tracks, followed by eight
	// ones on at least one track.
	register uint8_t zeros = 0, ones = 0;
	while (zeros < 24 || ones < 8) {

		for (w = 0xff; (((PINC & MASK_BOTH_TRACKS) ^ start) == 0) && w > 0; w--);

		if (w == 0) { // could not sync
			UDR0 = 3; // cancel pending PUT command
			interrupts();
			return 0;
		}

		_delay_us(2.0); // short delay to make sure track state has settled
		start = PINC & MASK_BOTH_TRACKS;

		if (IF1) _delay_us(6.50); else _delay_us(4.50);

		if (((end = PINC & MASK_BOTH_TRACKS) ^ start) == 0) {
			if (ones > 0) {
				ones = 0;
				zeros = 1;
			} else {
				zeros++;
			}
		} else {
			ones++;
		}
		start = end;
	}

	UDR0 = 0; // complete pending PUT command to go ahead

	while (true) {

		d = 0;

		for (bitCount = 4; bitCount > 0; bitCount--) {

			// wait for start of cycle, or end of block
			// skipped when coming here for the first time
			for (w = 0xff; (((PINC & MASK_BOTH_TRACKS) ^ start) == 0) && w > 0; w--);

			if (w == 0) { // end of block
				interrupts();
				return read;
			}

			_delay_us(2.0); // short delay to make sure track state has settled
			start = PINC & MASK_BOTH_TRACKS;        //   then take start reading
			                                        // and wait for end of cycle
			if (IF1) _delay_us(6.50); else _delay_us(4.50);
			// When a track has changed state compared to start of cycle at this
			// point, then it carries a 1 in this cycle, otherwise a 0.
			d = (d << 1) | ((end = PINC & MASK_BOTH_TRACKS) ^ start);   // store
			start = end;                               // prepare for next cycle
		}

		UDR0 = d; // send over serial
		read++;
	}
}

// ----------------------------------------------------------------- REPLAY ---

/*
	Replay one sector. Switching over into record mode or stopping of drive is
	continuously monitored, and if detected, returns immediately.
 */
void replay() {

	if (checkRecordingOrStop() || !isDriveReadable()) {
		return;
	}

	daemonCmdArgs(CMD_GET, activeDrive, 0, 0, 0);

	unsigned long start = millis();
	uint16_t rcv = daemonRcv(0);
	if (rcv == 0) {
		synced = millis() - start < RESYNC_THRESHOLD;
		return;
	}

	// header
	if (replayBlock(buffer + CMD_LENGTH, headerLengthMux)) {
		return;
	}

	headerGap = true;
	if (timerEnabled()) { // COMMS_CLK timer may be active at this point
		stopTimer();
	}
	setTimer(IF1 ? TIMER_HEADER_GAP_IF1 : TIMER_HEADER_GAP_QL, endHeaderGap);

	while (headerGap) {
		if (checkRecordingOrStop()) {
			return;
		}
		_delay_us(20.0);
	}

	// record - for QL, this can be two different lengths due to extra bytes
	// being sent during format
	replayBlock(buffer + CMD_LENGTH + headerLengthMux, rcv - headerLengthMux);
}

/*
	Get a sector from daemon and immediately reflect it back. For reliability
	testing. FIXME: validate
 */
void verify() {
	daemonCmdArgs(CMD_GET,
		lowByte(sectorLengthMux), highByte(sectorLengthMux), ' ', 0);
	daemonRcv(sectorLengthMux);
	// send back for verification
	daemonCmdArgs(CMD_VERIFY,
		lowByte(sectorLengthMux), highByte(sectorLengthMux), ' ',
		sectorLengthMux);
}

/*
	Indefinitely replays the given pattern for checking wave form with
	oscilloscope.
 */
void calibrate(uint8_t pattern) {
	calibration = true;
	for (int ix = 0; ix < BUF_LENGTH; ix++) {
		buffer[ix] = pattern;
	}
	setTracksToReplay();
	while (true) {
		replayBlock(buffer, BUF_LENGTH);
	}
}

/*
	Replay a block of bytes to the Interface 1/QL. Switching over to recording
	and stopping of drive is periodically checked. If either of those was
	detected, replay stops and true is returned. If replay was completed,
	false is returned.
 */
bool replayBlock(uint8_t* buf, uint16_t len) {

	noInterrupts();

	register uint8_t bitCount, d, tracks = MASK_BOTH_TRACKS;

	for (; len > 0; len--) {
		d = *buf;
		for (bitCount = 4; bitCount > 0; bitCount--) {
			tracks = ~tracks;            // tracks always flip at start of cycle
			PORTC = tracks | ~MASK_BOTH_TRACKS;                     // cycle end
			                                         // wait for middle of cycle
			if (IF1) _delay_us(5.40); else _delay_us(4.20);
			tracks = tracks ^ d;                   // flip track where data is 1
			PORTC = tracks | ~MASK_BOTH_TRACKS;         // write out track flips
			                                            // wait for end of cycle
			if (IF1) _delay_us(2.35); else _delay_us(1.15);
			// note that calibration must not be a compile time constant,
			// otherwise we'd get different timing due to optimizations
			if (checkRecordingOrStop() && !calibration) {
				interrupts();
				return true;
			}
			d = d >> 1;
		}
		buf++;
	}

	PORTC = 0x3f; // return tracks to idle level (HIGH) at cycle end

	interrupts();
	return false;
}

// --------------------------------------------------------- DEBUG MESSAGES ---

void debugMsg(uint8_t a, uint8_t b, uint8_t c) {
	msgBuffer[1] = a;
	msgBuffer[2] = b;
	msgBuffer[3] = c;
	message = true;
}

void debugFlush() {
	if (message) {
		msgBuffer[0] = CMD_DEBUG;
		Serial.write(msgBuffer, 4);
		Serial.flush();
		message = false;
	}
}

// --------------------------------------------------- DAEMON COMMUNICATION ---

//
void daemonSync() {
	driveState = DRIVE_STATE_UNKNOWN;
	while (true) {
		daemonCmd(IF1 ? IF1_HELLO : QL_HELLO);
		if (daemonRcvAck(10, 100, DAEMON_HELLO)) {
			lastPing = millis();
			synced = true;
			return;
		}
	}
}

//
void daemonPing() {
	if (millis() - lastPing < PING_INTERVAL) {
		return;
	}
	daemonCmd(DAEMON_PING);
	synced = daemonRcvAck(10, 5, DAEMON_PONG);
	lastPing = millis();
}

//
void daemonCmd(uint8_t cmd[]) {
	daemonCmdArgs(cmd[0], cmd[1], cmd[2], cmd[3], 0);
}

//
void daemonPendingCmd(uint8_t a, uint8_t b, uint8_t c) {
	buffer[0] = a;
	buffer[1] = b;
	buffer[2] = c;
	Serial.write(buffer, 3);
	Serial.flush();
}

//
void daemonCmdArgs(uint8_t cmd, uint8_t arg1, uint8_t arg2, uint8_t arg3,
	uint16_t bufferLen) {
	buffer[0] = cmd;
	buffer[1] = arg1;
	buffer[2] = arg2;
	buffer[3] = arg3;
	Serial.write(buffer, CMD_LENGTH + bufferLen);
	Serial.flush();
}

//
bool daemonRcvAck(uint8_t rounds, uint8_t wait, uint8_t exp[]) {

	for (int r = 0; r < rounds; r++) {
		if (Serial.available() < CMD_LENGTH) {
			delay(wait);

		} else {
			daemonRcv(CMD_LENGTH);
			for (int ix = 0; ix < CMD_LENGTH; ix++) {
				if (buffer[CMD_LENGTH + ix] != exp[ix]) {
					return false;
				}
			}
			return true;
		}
	}
	return false;
}

//
uint16_t daemonRcv(uint16_t bufferLen) {

	if (bufferLen == 0) { // unknown expected length, get from daemon
		if (daemonRcv(2) == 0) {
			return 0;
		};
		bufferLen = buffer[CMD_LENGTH]
			| (((uint16_t)buffer[CMD_LENGTH + 1]) << 8);
	}

	if (bufferLen > 0) {
		// don't overrun the buffer...
		uint16_t excess = bufferLen > PAYLOAD_LENGTH ? bufferLen - PAYLOAD_LENGTH : 0;
		uint16_t toRead = bufferLen - excess;
		if (Serial.readBytes(buffer + CMD_LENGTH, toRead) != toRead) {
			return 0;
		}
		// ...but still eat up all expected bytes coming in over serial
		uint8_t dummy[1];
		for (; excess > 0; excess--) {
			Serial.readBytes(dummy, 1);
		}
	}

	return bufferLen;
}

// ------------------------------------------------------------------ TIMER ---

void setTimer(int preload, TimerHandler h) {
	enableTimer(true, preload, h);
}

void stopTimer() {
	enableTimer(false, 0, NULL);
}

bool timerEnabled() {
	return TIMSK1 & (1<<TOIE1);
}

void enableTimer(bool on, int preload, TimerHandler h) {
	if (timerEnabled() == on) {
		return;
	}
	noInterrupts();
	timerHandler = h;
	if (on) {
		TCCR1A = 0;
		TCCR1B = 0;
		TIFR1 |= _BV(TOV1);   // clear the overflow interrupt flag
		TCNT1 = preload;
		TCCR1B |= (1<<CS12);  // 256 pre-scaler
		TIMSK1 |= (1<<TOIE1); // enable timer overflow interrupt
	} else {
		TIMSK1 &= (0<<TOIE1);
	}
	interrupts();
}

ISR(TIMER1_OVF_vect) {
	TimerHandler h = timerHandler;
	stopTimer();
	if (h != NULL) {
		h();
	}
}

// -------------------------------------------------------------------- LED ---

void ledReadFlip() {
	ledFlip(MASK_LED_READ);
}

void ledRead(bool active) {
	ledActivity(MASK_LED_READ, active, LED_RW_IDLE_ON);
}

void ledWriteFlip() {
	ledFlip(MASK_LED_WRITE);
}

void ledWrite(bool active) {
	ledActivity(MASK_LED_WRITE, active, LED_RW_IDLE_ON);
}

void ledActivity(uint8_t mask, bool active, bool idleOn) {
	(idleOn ? !active : active) ? ledOn(mask) : ledOff(mask);
}

void ledOn(uint8_t led_mask) {
	PORTB |= led_mask;
}

void ledOff(uint8_t led_mask) {
	PORTB &= (~led_mask);
}

void ledFlip(uint8_t led_mask) {
	PORTB ^= led_mask;
}
