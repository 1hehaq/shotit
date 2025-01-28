#!/bin/bash

RED='\033[1;91m'
GREEN='\033[1;92m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RESET='\033[0m'
ARROW="${GREEN}→${RESET}"
DOLLAR="${GREEN}\$${RESET}"

echo '
         __          __  _ __ 
   _____/ /_  ____  / /_(_) /_
  / ___/ __ \/ __ \/ __/ / __/
 (__  ) / / / /_/ / /_/ / /_  
/____/_/ /_/\____/\__/_/\__/  

                        @1hehaq
'

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

display_status() {
    local tool=$1
    if command_exists "$tool"; then
        echo -e "(${GREEN}installed${RESET})"
    else
        echo -e "(${YELLOW}not installed${RESET})"
    fi
}

install_requirements() {
    echo -e "\n[${BLUE}INF${RESET}] Installing basic requirements..."
    sudo apt-get update
    sudo apt-get install -y git curl wget build-essential python3 python3-pip golang rustc cargo nodejs npm
}

install_go_tool() {
    local repo=$1
    local binary=$2
    echo -e "\n[${BLUE}INF${RESET}] Installing $binary..."
    go install $repo@latest
}

install_python_tool() {
    local repo=$1
    local name=$2
    echo -e "\n[${BLUE}INF${RESET}] Installing $name..."
    git clone $repo /tmp/$name
    cd /tmp/$name
    pip3 install -r requirements.txt
    sudo python3 setup.py install
    cd - > /dev/null
    rm -rf /tmp/$name
}

update_go_tool() {
    local repo=$1
    local binary=$2
    echo -e "\n[${BLUE}INF${RESET}] Updating $binary..."
    if command_exists "$binary"; then
        go install -v "$repo@latest"
        echo -e "[${GREEN}SUCCESS${RESET}] Updated $binary"
    else
        echo -e "[${YELLOW}WARN${RESET}] $binary not found. Installing..."
        install_go_tool "$repo" "$binary"
    fi
}

update_git_tool() {
    local repo=$1
    local dir=$2
    local binary=$3
    
    if [ -d "$HOME/$dir" ]; then
        echo -e "\n[${BLUE}INF${RESET}] Updating $binary..."
        cd "$HOME/$dir"
        if git pull origin master; then
            echo -e "[${GREEN}SUCCESS${RESET}] Updated $binary"
        else
            echo -e "[${RED}ERR${RESET}] Failed to update $binary"
        fi
        cd - > /dev/null
    else
        echo -e "[${YELLOW}WARN${RESET}] $binary not found. Installing..."
        install_tool "$binary"
    fi
}

update_pip_tool() {
    local package=$1
    echo -e "\n[${BLUE}INF${RESET}] Updating $package..."
    if pip3 show "$package" >/dev/null 2>&1; then
        pip3 install --upgrade "$package"
        echo -e "[${GREEN}SUCCESS${RESET}] Updated $package"
    else
        echo -e "[${YELLOW}WARN${RESET}] $package not found. Installing..."
        pip3 install "$package"
    fi
}

echo -e "[${BLUE}INF${RESET}] OS: $(uname -s) [ARCH: $(uname -m)]\n"

echo -e "$ARROW go $(display_status go)"
echo -e "$ARROW python3 $(display_status python3)"
echo -e "$ARROW rust $(display_status cargo)"
echo -e "$ARROW nodejs $(display_status node)"

install_tool() {
    local tool=$1
    local skip_requirements=$2
    
    if [ "$skip_requirements" != "true" ]; then
        install_requirements
    fi
    
    case $tool in
        "sublist3r")
            install_python_tool "https://github.com/aboul3la/Sublist3r.git" "sublist3r"
            ;;
        "anew")
            install_go_tool "github.com/tomnomnom/anew" "anew"
            ;;
        "assetfinder")
            install_go_tool "github.com/tomnomnom/assetfinder" "assetfinder"
            ;;
        "gobuster")
            install_go_tool "github.com/OJ/gobuster/v3" "gobuster"
            ;;
        "puredns")
            install_go_tool "github.com/d3mondev/puredns/v2" "puredns"
            ;;
        "amass")
            install_go_tool "github.com/owasp-amass/amass/v4/..." "amass"
            ;;
        "findomain")
            curl -LO https://github.com/findomain/findomain/releases/latest/download/findomain-linux.zip
            unzip findomain-linux.zip
            chmod +x findomain
            sudo mv findomain /usr/local/bin/
            rm findomain-linux.zip
            ;;
        "dalfox")
            install_go_tool "github.com/hahwul/dalfox/v2" "dalfox"
            ;;
        "hakrawler")
            install_go_tool "github.com/hakluke/hakrawler" "hakrawler"
            ;;
        "subjs")
            install_go_tool "github.com/lc/subjs" "subjs"
            ;;
        "gospider")
            install_go_tool "github.com/jaeles-project/gospider" "gospider"
            ;;
        "ffuf")
            install_go_tool "github.com/ffuf/ffuf" "ffuf"
            ;;
        "gau")
            install_go_tool "github.com/lc/gau/v2/cmd/gau" "gau"
            ;;
        "waybackurls")
            install_go_tool "github.com/tomnomnom/waybackurls" "waybackurls"
            ;;
        "gf")
            install_go_tool "github.com/tomnomnom/gf" "gf"
            ;;
        "qsreplace")
            install_go_tool "github.com/tomnomnom/qsreplace" "qsreplace"
            ;;
        "massdns")
            git clone https://github.com/blechschmidt/massdns.git
            cd massdns
            make
            sudo make install
            cd ..
            rm -rf massdns
            ;;
        "paramspider")
            git clone https://github.com/devanshbatham/ParamSpider
            cd ParamSpider
            pip3 install -r requirements.txt
            sudo ln -s $(pwd)/paramspider.py /usr/local/bin/paramspider
            cd ..
            ;;
        "arjun")
            pip3 install arjun
            ;;
        "sqlmap")
            git clone --depth 1 https://github.com/sqlmapproject/sqlmap.git
            sudo ln -s $(pwd)/sqlmap/sqlmap.py /usr/local/bin/sqlmap
            ;;
        "all")
            install_requirements
            for tool in sublist3r anew assetfinder gobuster puredns amass findomain dalfox hakrawler subjs gospider ffuf gau waybackurls gf qsreplace massdns paramspider arjun sqlmap; do
                install_tool $tool "true"
            done
            ;;
        *)
            echo -e "[${RED}ERR${RESET}] Unknown tool: $tool"
            ;;
    esac
}

echo -e "\n${YELLOW}Install:${RESET}"
echo -e "$DOLLAR shotit install toolname       - Install a specific tool"
echo -e "$DOLLAR shotit install all            - Install all tools"
echo -e "$DOLLAR shotit update                 - Update all installed tools"
echo -e "$DOLLAR shotit list                   - List all available tools"

case $1 in
    "install")
        if [ "$2" = "all" ]; then
            install_tool "all"
        elif [ -n "$2" ]; then
            install_tool "$2"
        else
            echo -e "[${RED}ERR${RESET}] Please specify a tool to install"
        fi
        ;;
    "update")
        if [ "$2" = "all" ]; then
            echo -e "[${BLUE}INF${RESET}] Updating all installed tools..."
            
            echo -e "\n[${BLUE}INF${RESET}] Updating system packages..."
            sudo apt-get update && sudo apt-get upgrade -y
            
            if command_exists go; then
                echo -e "\n[${BLUE}INF${RESET}] Updating Go packages..."
                go get -u all
            fi
            
            if command_exists pip3; then
                echo -e "\n[${BLUE}INF${RESET}] Updating pip..."
                pip3 install --upgrade pip
            fi
            
            update_go_tool "github.com/tomnomnom/anew" "anew"
            update_go_tool "github.com/tomnomnom/assetfinder" "assetfinder"
            update_go_tool "github.com/OJ/gobuster/v3" "gobuster"
            update_go_tool "github.com/d3mondev/puredns/v2" "puredns"
            update_go_tool "github.com/owasp-amass/amass/v4/..." "amass"
            update_go_tool "github.com/hahwul/dalfox/v2" "dalfox"
            update_go_tool "github.com/hakluke/hakrawler" "hakrawler"
            update_go_tool "github.com/lc/subjs" "subjs"
            update_go_tool "github.com/jaeles-project/gospider" "gospider"
            update_go_tool "github.com/ffuf/ffuf" "ffuf"
            update_go_tool "github.com/lc/gau/v2/cmd/gau" "gau"
            update_go_tool "github.com/tomnomnom/waybackurls" "waybackurls"
            update_go_tool "github.com/tomnomnom/gf" "gf"
            update_go_tool "github.com/tomnomnom/qsreplace" "qsreplace"
            
            update_git_tool "https://github.com/aboul3la/Sublist3r.git" "Sublist3r" "sublist3r"
            update_git_tool "https://github.com/blechschmidt/massdns.git" "massdns" "massdns"
            update_git_tool "https://github.com/devanshbatham/ParamSpider" "ParamSpider" "paramspider"
            update_git_tool "https://github.com/sqlmapproject/sqlmap.git" "sqlmap" "sqlmap"
            
            update_pip_tool "arjun"

            if command_exists findomain; then
                echo -e "\n[${BLUE}INF${RESET}] Updating findomain..."
                curl -LO https://github.com/findomain/findomain/releases/latest/download/findomain-linux.zip
                unzip -o findomain-linux.zip
                chmod +x findomain
                sudo mv -f findomain /usr/local/bin/
                rm findomain-linux.zip
            fi
            
            echo -e "\n[${GREEN}SUCCESS${RESET}] All tools updated successfully!"
        elif [ -n "$2" ]; then
            case $2 in
                "go-tools")
                    echo -e "[${BLUE}INF${RESET}] Updating all go tools..."
                    go get -u all
                    ;;
                "pip-tools")
                    echo -e "[${BLUE}INF${RESET}] Updating all python tools..."
                    pip3 list --outdated | cut -d ' ' -f1 | xargs -n1 pip3 install -U
                    ;;
                *)
                    install_tool "$2"
                    ;;
            esac
        else
            echo -e "\n${YELLOW}Update:${RESET}"
            echo -e "$ARROW shotit update all        - Update all tools"
            echo -e "$ARROW shotit update go-tools   - Update only go tools"
            echo -e "$ARROW shotit update pip-tools  - Update only python tools"
            echo -e "$ARROW shotit update <tool>     - Update specific tool"
        fi
        ;;
    "list")
        echo
        echo -e "$ARROW sublist3r $(display_status sublist3r)"
        echo -e "$ARROW assetfinder $(display_status assetfinder)"
        echo -e "$ARROW amass $(display_status amass)"
        echo -e "$ARROW findomain $(display_status findomain)"
        echo -e "$ARROW puredns $(display_status puredns)"
        echo -e "$ARROW gobuster $(display_status gobuster)"
        echo -e "$ARROW hakrawler $(display_status hakrawler)"
        echo -e "$ARROW gospider $(display_status gospider)"
        echo -e "$ARROW ffuf $(display_status ffuf)"
        echo -e "$ARROW gau $(display_status gau)"
        echo -e "$ARROW waybackurls $(display_status waybackurls)"
        echo -e "$ARROW dalfox $(display_status dalfox)"
        echo -e "$ARROW sqlmap $(display_status sqlmap)"
        echo -e "$ARROW arjun $(display_status arjun)"
        echo -e "$ARROW anew $(display_status anew)"
        echo -e "$ARROW massdns $(display_status massdns)"
        echo -e "$ARROW gf $(display_status gf)"
        echo -e "$ARROW qsreplace $(display_status qsreplace)"
        echo -e "$ARROW paramspider $(display_status paramspider)"
        ;;
    *)
        echo -e "\n[${RED}ERR${RESET}] Invalid command. Use 'shotit list' to see available tools."
        ;;
esac
