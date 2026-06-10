import json
import time
import os
import sys
import subprocess
import tempfile

class Level:
    def __init__(self, path):
        if not os.path.exists(path):
            print(f"CRITICAL ERROR: Level file not found at {path}", file=sys.stderr)
            sys.exit(1)
        with open(path, "r") as f:
            self.data = json.load(f)
        self.id = self.data["id"]
        self.title = self.data["title"]
        self.description = self.data["description"]
        self.goal = self.data["goal"]
        self.allowed_builtins = self.data["allowed_builtins"]
        self.obstacles = self.data.get("obstacles", [])

class GameState:
    def __init__(self, level):
        self.tick = 0
        self.agent = {"x": 0, "y": 0}
        self.goal = level.goal
        self.status = "active"
        self.error_msg = ""
        self.level_metadata = {
            "id": level.id,
            "title": level.title,
            "description": level.description,
            "allowed_builtins": level.allowed_builtins,
            "obstacles": level.obstacles
        }

    def to_json(self):
        payload = {"tick": self.tick, "agent": self.agent, "goal": self.goal, "status": self.status}
        if self.error_msg:
            payload["error_msg"] = self.error_msg
        if self.tick == 0:
            payload["level"] = self.level_metadata
        return json.dumps(payload)

class Simulation:
    def __init__(self):
        self.level_path = os.getenv("PHX_LEVEL_PATH", "levels/level_1.json")
        self.level = Level(self.level_path)
        self.state = GameState(self.level)
        self.script_path = os.getenv("PHX_SCRIPT_PATH", "scripts/agent.py")
        self.last_mod = 0
        self.agent_process = None

    def check_collision(self, new_x, new_y):
        for obs in self.level.obstacles:
            if obs["x"] <= new_x < obs["x"] + obs["w"] and obs["y"] <= new_y < obs["y"] + obs["h"]:
                return True
        return False

    def handle_command(self, cmd):
        if cmd == "move_forward" and "move_forward" in self.level.allowed_builtins:
            new_x = self.state.agent["x"] + 1000
            if not self.check_collision(new_x, self.state.agent["y"]): self.state.agent["x"] = new_x
            else: self.state.status = "crashed"
        elif cmd == "move_backward" and "move_backward" in self.level.allowed_builtins:
            new_x = self.state.agent["x"] - 1000
            if not self.check_collision(new_x, self.state.agent["y"]): self.state.agent["x"] = new_x
            else: self.state.status = "crashed"
        elif cmd == "move_up" and "move_up" in self.level.allowed_builtins:
            new_y = self.state.agent["y"] - 1000
            if not self.check_collision(self.state.agent["x"], new_y): self.state.agent["y"] = new_y
            else: self.state.status = "crashed"
        elif cmd == "move_down" and "move_down" in self.level.allowed_builtins:
            new_y = self.state.agent["y"] + 1000
            if not self.check_collision(self.state.agent["x"], new_y): self.state.agent["y"] = new_y
            else: self.state.status = "crashed"
        elif cmd == "get_x":
            return str(self.state.agent["x"])
        elif cmd == "get_y":
            return str(self.state.agent["y"])
        elif cmd == "get_goal_x":
            return str(self.state.goal["x"])
        elif cmd == "get_goal_y":
            return str(self.state.goal["y"])
        return "OK"

    def execute_python(self):
        if not self.agent_process or self.agent_process.poll() is not None:
            # We wrap the user's Python code in a runner that communicates via stdin/stdout
            py_wrapper = f"""
import sys
import os

# API Stubs
def send_cmd(cmd):
    print(cmd)
    sys.stdout.flush()
    return sys.stdin.readline().strip()

def get_int(cmd):
    print(cmd)
    sys.stdout.flush()
    val = sys.stdin.readline().strip()
    return int(val) if val.lstrip('-').isdigit() else 0

def move_forward(): send_cmd("move_forward")
def move_backward(): send_cmd("move_backward")
def move_up(): send_cmd("move_up")
def move_down(): send_cmd("move_down")
def get_x(): return get_int("get_x")
def get_y(): return get_int("get_y")
def get_goal_x(): return get_int("get_goal_x")
def get_goal_y(): return get_int("get_goal_y")

# Execute User Code
try:
    with open("{os.path.abspath(self.script_path)}", "r") as f:
        code = f.read()
except FileNotFoundError:
    print("Script file temporarily unavailable.", file=sys.stderr)
    sys.exit(0)

try:
    # Provide the restricted API
    exec_globals = {{
        "__builtins__": {{"range": range}},
        "move_forward": move_forward,
        "move_backward": move_backward,
        "move_up": move_up,
        "move_down": move_down,
        "get_x": get_x,
        "get_y": get_y,
        "get_goal_x": get_goal_x,
        "get_goal_y": get_goal_y,
    }}
    exec(code, exec_globals)
except Exception as e:
    import traceback
    print("ERROR|" + repr(e))
    sys.stdout.flush()

print("DONE")
sys.stdout.flush()
"""
            with tempfile.NamedTemporaryFile(suffix=".py", delete=False, mode="w") as f:
                f.write(py_wrapper)
                wrapper_path = f.name
            
            # Start the process with pipes
            self.agent_process = subprocess.Popen(["python3", wrapper_path], stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True, bufsize=1)
        
        if not self.agent_process:
            return

        # Read one command per tick
        try:
            cmd = self.agent_process.stdout.readline().strip()
            if cmd == "DONE" or cmd == "":
                pass # Finished execution
            elif cmd.startswith("ERROR|"):
                self.state.status = "compile_error"
                self.state.error_msg = cmd[6:]
            else:
                response = self.handle_command(cmd)
                self.agent_process.stdin.write(response + "\\n")
                self.agent_process.stdin.flush()
        except Exception as e:
             self.state.status = "compile_error"
             self.state.error_msg = f"Python Run Error: {e}"

    def compile_and_run_cpp(self):
        # We wrap the user's C++ code in a runner that communicates via stdin/stdout
        cpp_wrapper = f"""
#include <iostream>
#include <string>

using namespace std;

// API Stubs
void send_cmd(string cmd) {{
    cout << cmd << endl;
    string response;
    cin >> response; // Wait for core acknowledgment
}}

int get_int(string cmd) {{
    cout << cmd << endl;
    int val;
    cin >> val;
    return val;
}}

void move_forward() {{ send_cmd("move_forward"); }}
void move_backward() {{ send_cmd("move_backward"); }}
void move_up() {{ send_cmd("move_up"); }}
void move_down() {{ send_cmd("move_down"); }}
int get_x() {{ return get_int("get_x"); }}
int get_y() {{ return get_int("get_y"); }}
int get_goal_x() {{ return get_int("get_goal_x"); }}
int get_goal_y() {{ return get_int("get_goal_y"); }}

// Include User Code
#include "{os.path.abspath(self.script_path)}"

int main() {{
    agent_main(); // User must define this
    cout << "DONE" << endl;
    return 0;
}}
"""
        with tempfile.NamedTemporaryFile(suffix=".cpp", delete=False, mode="w") as f:
            f.write(cpp_wrapper)
            wrapper_path = f.name
        
        bin_path = wrapper_path[:-4]
        compile_res = subprocess.run(["g++", "-std=c++11", wrapper_path, "-o", bin_path], capture_output=True, text=True)
        
        if compile_res.returncode != 0:
            self.state.status = "compile_error"
            self.state.error_msg = compile_res.stderr[:1000]
            return None

        # Start the process with pipes
        return subprocess.Popen([bin_path], stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True, bufsize=1)

    def execute_cpp(self):
        if not self.agent_process or self.agent_process.poll() is not None:
            self.agent_process = self.compile_and_run_cpp()
        
        if not self.agent_process:
            return

        # Read one command per tick
        # Note: In a robust setup, you'd use non-blocking IO or select. 
        # For this prototype, we'll let the C++ program block.
        try:
            cmd = self.agent_process.stdout.readline().strip()
            if cmd == "DONE" or cmd == "":
                pass # Finished execution
            else:
                response = self.handle_command(cmd)
                self.agent_process.stdin.write(response + "\\n")
                self.agent_process.stdin.flush()
        except Exception as e:
             print(f"C++ Run Error: {e}", file=sys.stderr)


    def compile_and_run_java(self):
        java_wrapper = f"""
import java.util.Scanner;

public class AgentRunner {{
    static Scanner scanner = new Scanner(System.in);

    public static void sendCmd(String cmd) {{
        System.out.println(cmd);
        scanner.nextLine(); // wait for ack
    }}

    public static int getInt(String cmd) {{
        System.out.println(cmd);
        String val = scanner.nextLine().trim();
        try {{
            return Integer.parseInt(val);
        }} catch (NumberFormatException e) {{
            return 0;
        }}
    }}

    public static void moveForward() {{ sendCmd("move_forward"); }}
    public static void moveBackward() {{ sendCmd("move_backward"); }}
    public static void moveUp() {{ sendCmd("move_up"); }}
    public static void moveDown() {{ sendCmd("move_down"); }}
    public static int getX() {{ return getInt("get_x"); }}
    public static int getY() {{ return getInt("get_y"); }}
    public static int getGoalX() {{ return getInt("get_goal_x"); }}
    public static int getGoalY() {{ return getInt("get_goal_y"); }}

    // Include User Code
    {open(self.script_path).read()}

    public static void main(String[] args) {{
        agent_main();
        System.out.println("DONE");
    }}
}}
"""
        with tempfile.TemporaryDirectory() as tmpdir:
            java_path = os.path.join(tmpdir, "AgentRunner.java")
            with open(java_path, "w") as f:
                f.write(java_wrapper)
            
            compile_res = subprocess.run(["javac", java_path], capture_output=True, text=True)
            if compile_res.returncode != 0:
                self.state.status = "compile_error"
                self.state.error_msg = compile_res.stderr[:1000]
                return None
            
            return subprocess.Popen(["java", "-cp", tmpdir, "AgentRunner"], stdin=subprocess.PIPE, stdout=subprocess.PIPE, text=True, bufsize=1)

    def execute_java(self):
        if not self.agent_process or self.agent_process.poll() is not None:
            self.agent_process = self.compile_and_run_java()
        
        if not self.agent_process:
            return

        try:
            cmd = self.agent_process.stdout.readline().strip()
            if cmd == "DONE" or cmd == "":
                pass
            else:
                response = self.handle_command(cmd)
                self.agent_process.stdin.write(response + "\\n")
                self.agent_process.stdin.flush()
        except Exception as e:
             print(f"Java Run Error: {e}", file=sys.stderr)

    def run(self):
        print(f"--- Chrysalis Multi-Lang Core: {self.level.title} ---", file=sys.stderr)
        
        while True:
            try:
                mod_time = os.path.getmtime(self.script_path)
                if mod_time > self.last_mod:
                    self.last_mod = mod_time
                    self.state = GameState(self.level)
                    if self.agent_process:
                        self.agent_process.terminate()
                        self.agent_process = None
            except FileNotFoundError:
                # The file is currently being swapped by the Godot client.
                # Skip this tick and try again next loop.
                time.sleep(0.1)
                continue
            except OSError:
                pass

            try:
                if self.state.status == "complete" or self.state.status == "crashed":
                    print(self.state.to_json())
                    sys.stdout.flush()
                    time.sleep(0.1)
                    continue

                # Route execution based on file extension
                if self.script_path.endswith(".py"):
                    self.execute_python()
                elif self.script_path.endswith(".cpp"):
                    self.execute_cpp()
                elif self.script_path.endswith(".java"):
                    self.execute_java()
                else:
                    print("Unsupported script language.", file=sys.stderr)

                if self.state.agent["x"] >= self.state.goal["x"] and self.state.agent["y"] >= self.state.goal["y"]:
                    self.state.status = "complete"

                print(self.state.to_json())
                sys.stdout.flush()

                self.state.tick += 1
                time.sleep(0.1)
            except BrokenPipeError:
                # Godot closed the pipe, terminate cleanly
                if self.agent_process:
                    self.agent_process.terminate()
                sys.exit(0)

if __name__ == "__main__":
    sim = Simulation()
    sim.run()
