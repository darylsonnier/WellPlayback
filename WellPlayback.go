package main

import (
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
)

var header = []string{"Hole Depth", "Bit Position", "Bit Weight", "Flow Out Percent", "Hook Load", "Pit Volume 1", "Pit Volume 2", "Pit Volume 3", "Pump Pressure", "Pump SPM 1", "Pump SPM 2", "Svy Azimuth", "Svy Depth", "Svy Inclination", "Top Drive RPM", "Top Drive Torque", "Toolface Grav", "Toolface Mag"}

//  XML Configuration
var xmlConfig XMLConfig

//  Well Constants
var standWeight, startDepth, pit1TotalVol, pit2TotalVol, pit3TotalVol, rpmMax, torqueMax, spmMax, bitWeightMax, flowMax, standLength, emptyBlockWeight float64
var totalStands int

//  Section Data
var depthOffset = rand.Float64() // feet -- random starting point
var holeDepth, bitPosition, hookLoad, pit1, pit2, pit3, pumpPressure, azimuth, inclination, grav, mag, rop, currentBitPosition, currentDepth float64
var bitWeight, flowOut, spm1, spm2, rpm, torque = 0.0, 0.0, 0.0, 0.0, 0.0, 0.0

//  Starting Data
var standNumber, lastStand = 1, 0

func main() {

	loadXMLConfig()
	holeDepth += startDepth
	if holeDepth > 3.0 {
		bitPosition = holeDepth - 3.0
	} else {
		bitPosition = holeDepth
	}

	rand.Seed(time.Now().UnixNano()) // Random seed based on RTC in nanoseconds

	file, err := os.Create("FEPlayback.csv") // Output CSV
	checkError("Cannot create file", err)
	defer file.Close()

	writer := csv.NewWriter(file) // CSV writer
	defer writer.Flush()

	err = writer.Write(header) // error
	checkError("Cannot write to file", err)

	fmt.Println("Creating FEPlayback.csv.")
	drill(writer)
	connection(writer)

	for standNumber < totalStands { // Loop through all stands
		drill(writer)
		connection(writer)
		standNumber += 1
	}
	fmt.Println("FEPlayback.csv created from configuration type", xmlConfig.WellType+".")
}

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func drill(drillWriter *csv.Writer) {
	prepSection()
	currentDepth = holeDepth
	rand.Seed(time.Now().UnixNano()) // Random seed
	var data = []string{flToString(holeDepth), flToString(bitPosition), flToString(bitWeight), flToString(flowOut), flToString(hookLoad), flToString(pit1), flToString(pit2), flToString(pit3), flToString(pumpPressure), flToString(spm1), flToString(spm2), flToString(azimuth), flToString(currentDepth), flToString(inclination), flToString(rpm), flToString(torque), flToString(grav), flToString(mag)}
	//drillWriter.Write(data)
	for i := 0; i < 20; i++ {
		hookLoad = (emptyBlockWeight + (float64(standNumber) * standWeight)) - bitWeight + noise(1.0)
		bitPosition += 0.01
		if rpm < rpmMax {
			rpm += rand.Float64() * 5.0
		} else {
			rpm += noise(1.0)
		}
		pit1 = pit1TotalVol + noise(3)
		pit2 = pit2TotalVol + noise(3)
		pit3 = pit3TotalVol + noise(3)
		data = []string{flToString(holeDepth), flToString(bitPosition), flToString(bitWeight), flToString(flowOut), flToString(hookLoad), flToString(pit1), flToString(pit2), flToString(pit3), flToString(pumpPressure), flToString(spm1), flToString(spm2), flToString(azimuth), flToString(currentDepth), flToString(inclination), flToString(rpm), flToString(torque), flToString(grav), flToString(mag)}
		drillWriter.Write(data)
	}
	for holeDepth < (currentDepth + standLength) {
		prepSection()
		bitPosition += rop
		if bitWeight < bitWeightMax {
			bitWeight += (rand.Float64() + 1.0)
		} else {
			bitWeight += noise(0.5)
		}
		if flowOut < flowMax {
			flowOut += (rand.Float64() + 1)
		} else {
			flowOut += noise(1.0)
		}
		hookLoad = (emptyBlockWeight + (float64(standNumber) * standWeight)) - bitWeight + noise(1.0)
		pit1 = pit1TotalVol + noise(3)
		pit2 = pit2TotalVol + noise(3)
		pit3 = pit3TotalVol + noise(3)
		if spm1 < spmMax {
			spm1 += rand.Float64() * 4.0
		} else {
			spm1 = spmMax + noise(0.5)
		}
		if spm2 < spmMax {
			spm2 += rand.Float64() * 4.0
		} else {
			spm2 = spmMax + noise(0.5)
		}
		pumpPressure = (spm1 + spm2) * 20.0
		if rpm < rpmMax {
			rpm += rand.Float64() * 5.0
		} else {
			rpm += noise(1.0)
		}
		if torque < torqueMax {
			torque += rand.Float64() * 5.0
		} else {
			torque += noise(3.0)
		}

		// Stupid conversion of numbers back into strings since Go's CSV support is like something from the 1960s.
		data = []string{flToString(holeDepth), flToString(bitPosition), flToString(bitWeight), flToString(flowOut), flToString(hookLoad), flToString(pit1), flToString(pit2), flToString(pit3), flToString(pumpPressure), flToString(spm1), flToString(spm2), flToString(azimuth), flToString(currentDepth), flToString(inclination), flToString(rpm), flToString(torque), flToString(grav), flToString(mag)}

		drillWriter.Write(data)

		// Sanity check
		if bitWeight > bitWeightMax {
			bitWeight = bitWeightMax + noise(1.0)
		}
		if flowOut > flowMax {
			flowOut = flowMax + noise(1.0)
		}
		if rpm > rpmMax {
			rpm = rpmMax + noise(1.0)
		}
		if torque > torqueMax {
			torque = torqueMax + noise(3.0)
		}
		if holeDepth < bitPosition {
			holeDepth = bitPosition
		}
	}
}

func connection(connWriter *csv.Writer) {
	rand.Seed(time.Now().UnixNano()) // Random seed
	var data = []string{flToString(holeDepth), flToString(bitPosition), flToString(bitWeight), flToString(flowOut), flToString(hookLoad), flToString(pit1), flToString(pit2), flToString(pit3), flToString(pumpPressure), flToString(spm1), flToString(spm2), flToString(azimuth), flToString(currentDepth), flToString(inclination), flToString(rpm), flToString(torque), flToString(grav), flToString(mag)}

	currentBitPosition = bitPosition
	for (currentBitPosition - bitPosition) < 3.0 {
		bitPosition -= 0.1
		bitWeight = 0
		hookLoad = float64(standNumber)*standWeight + emptyBlockWeight + noise(1.0)
		// Stupid conversion of numbers back into strings since Go's CSV support is like something from the 1960s.
		data = []string{flToString(holeDepth), flToString(bitPosition), flToString(bitWeight), flToString(flowOut), flToString(hookLoad), flToString(pit1), flToString(pit2), flToString(pit3), flToString(pumpPressure), flToString(spm1), flToString(spm2), flToString(azimuth), flToString(currentDepth), flToString(inclination), flToString(rpm), flToString(torque), flToString(grav), flToString(mag)}
		connWriter.Write(data)
	}

	for i := 0; i < 60; {
		rand.Seed(time.Now().UnixNano()) // Random seed
		flowOut -= rand.Float64() + 1.0
		spm1 -= rand.Float64() + 1.0
		spm2 -= rand.Float64() + 1.0
		pumpPressure = (spm1 + spm2) * 20
		rpm -= rand.Float64() * 5.0
		torque -= rand.Float64() * 10.0
		hookLoad = float64(standNumber)*standWeight + emptyBlockWeight + noise(1.0)
		if flowOut < 0 {
			flowOut = 0.0
		}
		if spm1 < 0 {
			spm1 = 0
		}
		if spm2 < 0 {
			spm2 = 0
		}
		if pumpPressure < 0 {
			pumpPressure = 0
		}
		if rpm < 0 {
			rpm = 0
		}
		if torque < 0 {
			torque = 0
		}
		i += 1
		data = []string{flToString(holeDepth), flToString(bitPosition), flToString(bitWeight), flToString(flowOut), flToString(hookLoad), flToString(pit1), flToString(pit2), flToString(pit3), flToString(pumpPressure), flToString(spm1), flToString(spm2), flToString(azimuth), flToString(currentDepth), flToString(inclination), flToString(rpm), flToString(torque), flToString(grav), flToString(mag)}
		connWriter.Write(data)
	}

	// Stupid conversion of numbers back into strings since Go's CSV support is like something from the 1960s.
	//data = []string{flToString(holeDepth), flToString(bitPosition), flToString(bitWeight), flToString(flowOut), flToString(hookLoad), flToString(pit1), flToString(pit2), flToString(pit3), flToString(pumpPressure), flToString(spm1), flToString(spm2), flToString(azimuth), flToString(currentDepth), flToString(inclination), flToString(rpm), flToString(torque), flToString(grav), flToString(mag)}
	//connWriter.Write(data)

}

func flToString(input float64) string {
	retValue := fmt.Sprintf("%f", input)
	return retValue
}

func noise(noiseLevel float64) float64 {
	retValue := rand.Float64()*(noiseLevel-(-1*noiseLevel)) + (-1 * noiseLevel)
	return retValue
}

func loadXMLConfig() {

	byteValue, err := ioutil.ReadFile("config.xml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	xml.Unmarshal(byteValue, &xmlConfig)

	pit1TotalVol = xmlConfig.WellConstants.Pit1TotalVol
	pit2TotalVol = xmlConfig.WellConstants.Pit2TotalVol
	pit3TotalVol = xmlConfig.WellConstants.Pit3TotalVol
	standLength = xmlConfig.WellConstants.StandLength
	totalStands = xmlConfig.WellConstants.TotalStands
	startDepth = xmlConfig.WellConstants.StartDepth
	emptyBlockWeight = xmlConfig.WellConstants.EmptyBlock
	fmt.Println("Loaded config.xml.")
}

func prepSection() {
	index := 0
	for i, section := range xmlConfig.Sections.Sections {
		if holeDepth < section.Depth {
			index = i
			break
		}
	}
	//fmt.Println("Section", (index + 1), xmlConfig.Sections.Sections[index].BitWeightMax)
	// Set maximums
	bitWeightMax = xmlConfig.Sections.Sections[index].BitWeightMax
	flowMax = xmlConfig.Sections.Sections[index].FlowMax
	spmMax = xmlConfig.Sections.Sections[index].SpmMax
	rpmMax = xmlConfig.Sections.Sections[index].RpmMax
	torqueMax = xmlConfig.Sections.Sections[index].TorqueMax
	// Set section conditions
	rop = xmlConfig.Sections.Sections[index].Rop / 3600.0
	rop += rop + noise(0.01*rop)
	azimuth = xmlConfig.Sections.Sections[index].Azimuth
	inclination = xmlConfig.Sections.Sections[index].Inclination
	mag = xmlConfig.Sections.Sections[index].Mag
	grav = xmlConfig.Sections.Sections[index].Grav
	standWeight = xmlConfig.Sections.Sections[index].StandWeight
}

type XMLConfig struct {
	XMLName       xml.Name         `xml:"config"`
	WellType      string           `xml:"wellType,attr"`
	WellConstants XMLWellConstants `xml:"wellConstants"`
	Sections      XMLSections      `xml:"sections"`
}

type XMLWellConstants struct {
	XMLName      xml.Name `xml:"wellConstants"`
	Pit1TotalVol float64  `xml:"pit1TotalVol"`
	Pit2TotalVol float64  `xml:"pit2TotalVol"`
	Pit3TotalVol float64  `xml:"pit3TotalVol"`
	StandLength  float64  `xml:"standLength"`
	TotalStands  int      `xml:"totalStands"`
	StartDepth   float64  `xml:"startDepth"`
	EmptyBlock   float64  `xml:"emptyBlock"`
}

type XMLSections struct {
	XMLName  xml.Name     `xml:"sections"`
	Sections []XMLSection `xml:"section"`
}

type XMLSection struct {
	XMLName      xml.Name `xml:"section"`
	Depth        float64  `xml:"depth,attr"`
	Rop          float64  `xml:"rop,attr"`
	StandWeight  float64  `xml:"standWeight,attr"`
	RpmMax       float64  `xml:"rpmMax,attr"`
	TorqueMax    float64  `xml:"torqueMax,attr"`
	SpmMax       float64  `xml:"spmMax,attr"`
	BitWeightMax float64  `xml:"bitWeightMax,attr"`
	FlowMax      float64  `xml:"flowMax,attr"`
	Azimuth      float64  `xml:"azimuth,attr"`
	Inclination  float64  `xml:"inclination,attr"`
	Mag          float64  `xml:"mag,attr"`
	Grav         float64  `xml:"grav,attr"`
}
