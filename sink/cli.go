package sink

import "fmt"

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	colorReset  = "\033[0m"
)

func Banner(componentName string) {
	fmt.Println(string(colorBlue))
	fmt.Println("                                                                                   ")
	fmt.Println("  sOOsOsOOs      sSSs   .I   .N_NNNn   sdTT_TTTTTTbs  .A_AAAa     .M_   _M.    .I  ")
	fmt.Println(" d%%SP~YS%%b    d%%SP  .SS  .SS~YS%%b  YSSS~S%SSSSSP .SS~SSSSS   .SS~S*S~SS.  .SS  ")
	fmt.Println("d%S'     `S%b  d%S'    S%S  S%S   `S%b     `S%S      S%S   SSSS  S%S \\%/ S%S  S%S  ")
	fmt.Println("S%S   ~   S%S  S%|     S%S  S%S    S%S      S%S      S%S    S%S  S%S  |  S%S  S%S  ")
	fmt.Println("S&S ( O ) S&S  S&S     S&S  S%S    S&S      S&S      S%S SSSS%S  S%S     S%S  S&S  ")
	fmt.Println("S&S   ~   S&S  Y&Ss    S&S  S&S    S&S      S&S      S&S  SSS%S  S&S     S&S  S&S  ")
	fmt.Println("S&S       S&S  `S&&S   S&S  S&S    S&S      S&S      S&S    S&S  S&S     S&S  S&S  ")
	fmt.Println("S*b       d*S    dl*S  S*S  S*S    S*S      S*S      S*S    S&S  S*S     S*S  S*S  ")
	fmt.Println(" SSSbs_sdSSS   sSS*S   S*S  S*S    S*S      S*S      S*S    S*S  S*S     S*S  S*S  ")
	fmt.Println("  YSSP~YSSY    YSS'    S*S  S*S    SSS      S*S      SSS    S*S  SSS     S*S  S*S  ")
	fmt.Println("                       SP   SP              SP              SP           SP   SP   ")
	fmt.Println("                                                                                   ")
	fmt.Println(string(colorRed))
	fmt.Println("                                                                             v1.0.0")
	fmt.Println("                                                        OSINTAMI Project - @eslowan")
	fmt.Println("                                                          " + componentName)
	fmt.Println("                                                                                   ")
	fmt.Println(string(colorReset))
}
