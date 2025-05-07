package Root_System_Analysis

import (
	"encoding/xml"
	"fmt"
	"os"
	"strconv"
	"strings"
	"Project/component/Public_Data" 
	"Project/component/Subsystem_Analysis" 
	"path/filepath"
)

// Analyze the root XML file
func AnalyzeRootXML(xmlPath string, modelName string) error {
	file, err := os.Open(xmlPath)
	if err != nil {
		return fmt.Errorf("Failed to read XML file: %v", err)
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	var (
		inSubsystem      bool
		hasValidPorts    bool
		currentSystemRef string
		blockName        string
		blockSID         int
		portCountsIn     int
		portCountsOut    int
		portsParam       string
	)

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Block" {
				// reset state
				inSubsystem = false
				hasValidPorts = false
				currentSystemRef = ""
				blockName = ""
				blockSID = 0
				portCountsIn = 0
				portCountsOut = 0
				portsParam = ""

				// Collect basic attributes
				for _, attr := range se.Attr {
					switch attr.Name.Local {
					case "Name":
						blockName = attr.Value
					case "SID":
						if sid, err := strconv.Atoi(attr.Value); err == nil {
							blockSID = sid
						}
					case "BlockType":
						if attr.Value == "SubSystem" {
							inSubsystem = true
						}
					}
				}
			}

			if inSubsystem {
				switch se.Name.Local {
				case "PortCounts":
					hasValidPorts = false
					for _, attr := range se.Attr {
						switch attr.Name.Local {
						case "in":
							if in, err := strconv.Atoi(attr.Value); err == nil {
								portCountsIn = in
								hasValidPorts = true
							}
						case "out":
							if out, err := strconv.Atoi(attr.Value); err == nil {
								portCountsOut = out
								hasValidPorts = true
							}
						}
					}
				case "P":
					for _, attr := range se.Attr {
						if attr.Name.Local == "Name" && attr.Value == "Ports" {
							portsParam = "processing"
						}
					}
				case "System":
					for _, attr := range se.Attr {
						if attr.Name.Local == "Ref" {
							currentSystemRef = attr.Value
						}
					}
				}
			}

		case xml.CharData:
			if inSubsystem && portsParam == "processing" {
				portsStr := strings.TrimSpace(string(se))
				if in, out, err := parsePortsParam(portsStr); err == nil {
					portCountsIn = in
					portCountsOut = out
					hasValidPorts = true
				}
				portsParam = ""
			}

		case xml.EndElement:
			if se.Name.Local == "Block" && inSubsystem {
				if hasValidPorts && currentSystemRef != "" {
					// Create an effective system instance
					currentSystem := Public_Data.NewSystemFromBlock(
						blockName,
						blockSID,
						portCountsIn,
						portCountsOut,
						true,
					)
					newPath := generateNewPath(xmlPath, currentSystemRef)
					modelName := filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(xmlPath))))
					Subsystem_Analysis.AnalyzeSubSystemXML(newPath, currentSystem, modelName)
				}
				inSubsystem = false
			}
		}
	}
	return nil
}

// Analyze Ports parameters
func parsePortsParam(portsStr string) (int, int, error) {
	trimmed := strings.Trim(strings.TrimSpace(portsStr), "[]")
	if trimmed == "" {
		return 0, 0, fmt.Errorf("invalid ports format")
	}

	parts := strings.Split(trimmed, ",")
	if len(parts) < 2 {
		return 0, 0, fmt.Errorf("invalid ports format")
	}

	in, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	out, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))

	if err1 != nil || err2 != nil {
		return 0, 0, fmt.Errorf("parse ports failed")
	}

	return in, out, nil
}

// Path generation function
func generateNewPath(originalPath, systemRef string) string {
	return filepath.Join(
		filepath.Dir(originalPath),
		systemRef+".xml",
	)
}