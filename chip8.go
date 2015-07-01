package main

import (
        "fmt"
        "math/rand"  
        "time"
        "io/ioutil"
        "github.com/veandco/go-sdl2/sdl"
        "os"
)

const REGISTER_COUNT = 16 
const SPRITE_WIDTH = 8
const MEM_FONT_OFFSET = 0
const FONT_SIZE = 5
const START_PC = 0x200

const CLEAR_SCREEN = 0x00E0 
const SUBROUTINE_RETURN = 0x00EE
const CALL_PROGRAM = 0x0000

const OPCODE_MASK = 0xF000
const ADDRESS_MASK = 0x0FFF
const X_REGISTER_MASK = 0x0F00
const Y_REGISTER_MASK = 0x00F0
const CONSTANT_8_MASK = 0x00FF
const CONSTANT_4_MASK = 0x000F
const OPERATION_TYPE_MASK = 0x000F
const REGISTER_CARRY = 0xF

const OPCODE_BODY_MASK = 0x0FFF
const JUMP = 0x1000
const SUBROUTINE_CALL = 0x2000
const SKIP_EQUAL = 0x3000
const SKIP_N_EQUAL = 0x4000
const SKIP_REGS_EQUAL = 0x5000 
const MOV = 0x6000
const ADD = 0x7000
const OPERATION = 0x8000
const SKIP_REGS_N_EQUAL = 0x9000
const SET_ADDRESS = 0xA000 
const JUMP_OFFSET = 0xB000
const RANDOM = 0xC000
const DRAW = 0xD000
const KEY_EVENT = 0xE000

const SET_EQUAL_OP = 0x0
const OR_OP = 0x1
const AND_OP = 0x2
const XOR_OP = 0x3
const ADD_OP = 0x4
const SUB_OP = 0x5
const SHIFT_R_OP = 0x6
const SUB_INVERSE_OP = 0x7
const SHIFT_L_OP = 0xe

const KEY_PRESS_SKIP = 0x009E
const KEY_N_PRESS_SKIP = 0x00A1 

const GET_DELAY = 0x0007
const KEY_PRESS_CODE = 0x000A
const SET_DELAY = 0x0015
const SET_SOUND = 0x0018
const INCREMENT_ADDRESS = 0x001E
const GET_CHAR = 0x0029
const SET_MEMORY_DECIMAL = 0x0033
const SET_MEMORY = 0x0055
const GET_MEMORY = 0x0065

const (
  WIN_WIDTH = 640
  WIN_HEIGHT = 480
  CHIP_SCREEN_WIDTH = 64
  CHIP_SCREEN_HEIGHT = 32
)

const (
  WHITE = 0xFFFFFF
  BLACK = 0x000000
)

type Chip8 struct {

  opcode uint16
  memory [4096]byte
  v [REGISTER_COUNT]byte

  i uint16
  pc uint16

  gfx [CHIP_SCREEN_WIDTH * CHIP_SCREEN_HEIGHT]byte

  delay_timer byte
  sound_timer byte

  stack [16]uint16
  sp uint16

  key [16]byte

  drawFlag bool
  event sdl.Event
  quit bool

  surface *sdl.Surface

  key_mapping map[sdl.Keycode]byte  //key mapping between SDL and hexadecimal keyboard
}

type InvalidOpcodeError struct {
  opcode uint16
}

func (error *InvalidOpcodeError) Error() string {
  return fmt.Sprintf("Invalid Opcode: %x", error.opcode)
}


func (chip *Chip8) init() {
  rand.Seed(time.Now().UnixNano())

  chip.opcode = 0
  chip.pc = START_PC
  chip.i = 0
  chip.sp = 0

 /* chip.key_mapping = map[sdl.Keycode]byte{
    sdl.K_SPACE: 0x0,
    sdl.K_ESCAPE: 0x1,
    sdl.K_LCTRL: 0x2,
    sdl.K_w: 0x3,
    sdl.K_a: 0x4,
    sdl.K_s: 0x5,
    sdl.K_d: 0x6,
    sdl.K_e: 0x7,
    sdl.K_q: 0x8,
    sdl.K_r: 0x9,
    sdl.K_LSHIFT: 0xa,
    sdl.K_TAB: 0xb,
    sdl.K_RETURN: 0xc,
    sdl.K_1: 0xd,
    sdl.K_2: 0xe,
    sdl.K_3: 0xf,
  }*/

chip.key_mapping = map[sdl.Keycode]byte{
    sdl.K_x: 0x0,
    sdl.K_1: 0x1,
    sdl.K_2: 0x2,
    sdl.K_3: 0x3,
    sdl.K_q: 0x4,
    sdl.K_w: 0x5,
    sdl.K_e: 0x6,
    sdl.K_a: 0x7,
    sdl.K_s: 0x8,
    sdl.K_d: 0x9,
    sdl.K_z: 0xa,
    sdl.K_c: 0xb,
    sdl.K_4: 0xc,
    sdl.K_r: 0xd,
    sdl.K_f: 0xe,
    sdl.K_v: 0xf,
  }

  fontset := []byte{
    0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
    0x20, 0x60, 0x20, 0x20, 0x70, // 1
    0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
    0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
    0x90, 0x90, 0xF0, 0x10, 0x10, // 4
    0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
    0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
    0xF0, 0x10, 0x20, 0x40, 0x40, // 7
    0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
    0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
    0xF0, 0x90, 0xF0, 0x90, 0x90, // A
    0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
    0xF0, 0x80, 0x80, 0x80, 0xF0, // C
    0xE0, 0x90, 0x90, 0x90, 0xE0, // D
    0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
    0xF0, 0x80, 0xF0, 0x80, 0x80,  // F
  }

  //fill memory with fontset
  for i, c := range fontset {
    chip.memory[MEM_FONT_OFFSET + i] = c 
  }


}

func (c *Chip8) loadGame(game string) {
  fmt.Println("Loading: ", game)

  data, err := ioutil.ReadFile(game); 

  if err != nil {
    fmt.Println("Unable to find ", game)
    panic(err)
  }
  //fill memory with program
  for i := 0; i < len(data); i++ {
    c.memory[i + START_PC] = data[i];
  }
}




func operationCycle(chip *Chip8, opcode uint16) error {

  reg_y := (opcode & Y_REGISTER_MASK) >> 4 
  reg_x := (opcode & X_REGISTER_MASK) >> 8

  switch opcode &  OPERATION_TYPE_MASK {

    case SET_EQUAL_OP:
      chip.v[reg_x] = chip.v[reg_y]
      break

    case OR_OP:
      chip.v[reg_x] |= chip.v[reg_y]
      break

    case AND_OP:
      chip.v[reg_x] &= chip.v[reg_y]
      break

    case XOR_OP:
      chip.v[reg_x] ^= chip.v[reg_y]
      break

    case ADD_OP:
      if chip.v[reg_y] > (0xFF - chip.v[reg_x]) { //check overflow for 255
        chip.v[REGISTER_CARRY] = 1
      } else {
        chip.v[REGISTER_CARRY] = 0
      }
      chip.v[reg_x] += chip.v[reg_y]
      break

    case SUB_OP:

      if chip.v[reg_y] > chip.v[reg_x] { //check borrow for 255
        chip.v[REGISTER_CARRY] = 0
      } else {
        chip.v[REGISTER_CARRY] = 1
      }
      chip.v[reg_x] -= chip.v[reg_y]
      break

    case SHIFT_R_OP:
      chip.v[REGISTER_CARRY] = chip.v[reg_x] & 0x1 
      chip.v[reg_x] >>= 1
      break

    case SUB_INVERSE_OP:
      if chip.v[reg_x] > chip.v[reg_y] { //check borrow for 255
        chip.v[REGISTER_CARRY] = 0
      } else {
        chip.v[REGISTER_CARRY] = 1
      }
      chip.v[reg_x] = chip.v[reg_y] - chip.v[reg_x]
      break

    case SHIFT_L_OP:
      chip.v[REGISTER_CARRY] = chip.v[reg_x] & 0x80 
      chip.v[reg_x] <<= 1
      break

    default:
      return &InvalidOpcodeError{opcode}
  }
  return nil
}

func keyEventCycle(chip *Chip8, opcode uint16) error{
  reg_x := (opcode & X_REGISTER_MASK) >> 8
    
  switch opcode & CONSTANT_8_MASK {
  case KEY_PRESS_SKIP:
    if chip.key[chip.v[reg_x]] != 0 {
      chip.pc += 2
    }
    chip.pc += 2
    break

  case KEY_N_PRESS_SKIP:
    if chip.key[chip.v[reg_x]] == 0 {
      chip.pc += 2
    }
    chip.pc += 2
    break

  default:
    return &InvalidOpcodeError{opcode}
  }
  return nil
}

func otherOptionsCycle(chip *Chip8, opcode uint16) error{

  reg_x := (opcode & X_REGISTER_MASK) >> 8
  switch opcode & CONSTANT_8_MASK {

  case GET_DELAY:
    chip.v[reg_x] = chip.delay_timer
    break

  case KEY_PRESS_CODE:
    //bloked on waitEvent()
    chip.v[reg_x] = chip.waitEvent()
    break

  case SET_DELAY:
    chip.delay_timer = chip.v[reg_x]
    break

  case SET_SOUND:
    chip.sound_timer = chip.v[reg_x]
    break

  case INCREMENT_ADDRESS:
    if chip.i + uint16(chip.v[reg_x]) > 0xFFFF {
      chip.v[REGISTER_CARRY] = 1
    } else {
      chip.v[REGISTER_CARRY] = 0
    }//according to wiki overflow happens at 0xFFF, but I can go up to 0xFFFF
    chip.i += uint16(chip.v[reg_x])
    break

  case GET_CHAR:
    chip.i = uint16(chip.v[reg_x]) * FONT_SIZE
    break

  case SET_MEMORY_DECIMAL:
    val := chip.v[reg_x]
    chip.memory[chip.i] = val / 100
    chip.memory[chip.i + 1] = (val/10) % 10
    chip.memory[chip.i + 2] = (val%100) % 10
    break

  case SET_MEMORY:
    for i := 0; i <= int(reg_x); i++ {
      chip.memory[int(chip.i) + i] = chip.v[i]
    }
    break

  case GET_MEMORY:
    for i := 0; i <= int(reg_x); i++ {
      chip.v[i] = chip.memory[int(chip.i) + i] 
    }
    break

  default:
    return &InvalidOpcodeError{opcode}
  }
  chip.pc +=2
  return nil  
}

func (chip *Chip8) emulateCycle() error {

  var opcode uint16  = uint16(chip.memory[chip.pc]) << 8 | uint16(chip.memory[chip.pc + 1])
  fmt.Printf("%x\n", opcode)
  switch opcode & OPCODE_MASK {

  case JUMP:
    chip.pc = opcode & ADDRESS_MASK
    break

  case SUBROUTINE_CALL:
    chip.stack[chip.sp] = chip.pc + 2//save the next program counter
    chip.sp++
    chip.pc = opcode & ADDRESS_MASK
    break

  case SKIP_EQUAL:
    reg_x := (opcode & X_REGISTER_MASK) >> 8
    if uint16(chip.v[reg_x]) == (opcode & CONSTANT_8_MASK) {
      chip.pc += 2
    }
    chip.pc += 2
    break

  case SKIP_N_EQUAL:
    reg_x := (opcode & X_REGISTER_MASK) >> 8
    if uint16(chip.v[reg_x]) != (opcode & CONSTANT_8_MASK) {
      chip.pc += 2
    }
    chip.pc += 2
    break

  case SKIP_REGS_EQUAL:
    reg_x := (opcode & X_REGISTER_MASK) >> 8
    reg_y := (opcode & Y_REGISTER_MASK) >> 4

    if chip.v[reg_x] == chip.v[reg_y] {
         chip.pc += 2
    }
    chip.pc += 2

    break

  case MOV:
    reg_x := (opcode & X_REGISTER_MASK) >> 8
    chip.v[reg_x] = byte(opcode & CONSTANT_8_MASK)
    chip.pc += 2
    break

  case ADD:
    reg_x := (opcode & X_REGISTER_MASK) >> 8
    chip.v[reg_x] += byte(opcode & CONSTANT_8_MASK)
    chip.pc += 2
    break

  case OPERATION:
    if err := operationCycle(chip, opcode); err != nil  {
      return err
    }
    chip.pc += 2
    break

  case SKIP_REGS_N_EQUAL:
    reg_x := (opcode & X_REGISTER_MASK) >> 8
    reg_y := (opcode & Y_REGISTER_MASK) >> 4
    if chip.v[reg_x] != chip.v[reg_y] {
      chip.pc += 2
    }
    chip.pc += 2
    break

  case SET_ADDRESS:
    chip.i = opcode & OPCODE_BODY_MASK
    chip.pc +=2
    break

  case JUMP_OFFSET:
    chip.pc += (opcode & ADDRESS_MASK) + uint16(chip.v[0])
    break

  case RANDOM:
    reg_x := (opcode & X_REGISTER_MASK) >> 8
    chip.v[reg_x] = byte(uint16(rand.Intn(256)) & (opcode & CONSTANT_8_MASK)) //only 256 combinations possible
    chip.pc += 2
    break

  case DRAW:
    reg_x := (opcode & X_REGISTER_MASK) >> 8
    reg_y := (opcode & Y_REGISTER_MASK) >> 4 
    height := opcode & CONSTANT_4_MASK

    x := chip.v[reg_x]
    y := chip.v[reg_y]

    chip.v[REGISTER_CARRY] = 0
    for y_off := 0 ; uint16(y_off) < height && y_off + int(y) < CHIP_SCREEN_HEIGHT; y_off++ {

      row := chip.memory[int(chip.i) + y_off]
      
      for x_off := 0 ; x_off < SPRITE_WIDTH && x_off + int(x) < CHIP_SCREEN_WIDTH; x_off++ {
        

        if (uint16(row) & uint16(0x80 >> uint(x_off))) == 0 {
          continue
        }
        pixel := chip.gfx[(int(y) + y_off) * CHIP_SCREEN_WIDTH + (int(x) + x_off)] 
        if pixel == 1{
          chip.v[REGISTER_CARRY] = 1
        }
        chip.gfx[(int(y) + y_off) * CHIP_SCREEN_WIDTH + (int(x) + x_off)] ^= 1
      }
    }
    chip.drawFlag = true
    chip.pc += 2
    break

  case KEY_EVENT:
    if err := keyEventCycle(chip, opcode); err != nil {
      return err
    }
    break

  case 0xF000:
    if err := otherOptionsCycle(chip, opcode); err != nil {
      return err
    }
    break


  case 0x0000 :{
    switch opcode {

    case CLEAR_SCREEN:
      chip.clearScreen()
      chip.pc += 2
      break;

    case CALL_PROGRAM:
      //??
      break;

    case SUBROUTINE_RETURN:
      if chip.sp > 0{
        chip.sp--
        chip.pc = chip.stack[chip.sp] 
      }
      break;

    default:
      return &InvalidOpcodeError{opcode}
    }
    break
  }
  default:
    return &InvalidOpcodeError{opcode}
  }


  if chip.delay_timer > 0 {
    chip.delay_timer--
  }

  if chip.sound_timer > 0 {
    if chip.sound_timer == 1 {
      fmt.Println("BEEP")
    }
    chip.sound_timer--
  }


  return nil
}

func RectAt(index int) sdl.Rect {
    y := index / CHIP_SCREEN_WIDTH
    x := index % CHIP_SCREEN_WIDTH
    //if x > 63 || x < 0 || y > 32 || y < 0 {fmt.Println("Wrong values %d %d",x,y)}
    return sdl.Rect{int32(x*(WIN_WIDTH/CHIP_SCREEN_WIDTH)), 
                    int32(y*(WIN_HEIGHT/CHIP_SCREEN_HEIGHT)),
                    WIN_WIDTH/CHIP_SCREEN_WIDTH, 
                    WIN_HEIGHT/CHIP_SCREEN_HEIGHT}    
}


func (c *Chip8) drawGraphics() {
  for i, color := range c.gfx {
    rect := RectAt(i)
    if color == 1{
      c.surface.FillRect(&rect, WHITE)
    }else {
      c.surface.FillRect(&rect, BLACK)
    }
  }
}

func (c *Chip8) clearScreen() {
  for i, _ := range c.gfx {
    c.gfx[i] = 0
  }
}

func (c *Chip8) setKeys() {
  c.surface.FillRect(nil, BLACK)
  for c.event = sdl.PollEvent(); c.event != nil;  c.event = sdl.PollEvent() {
    switch event := c.event.(type) {

    case *sdl.KeyDownEvent:
      if code, found := c.key_mapping[event.Keysym.Sym]; found {
        c.key[code] = 1
      }
      break

    case *sdl.KeyUpEvent:
      if code, found := c.key_mapping[event.Keysym.Sym]; found {
        c.key[code] = 0
      }
      break

    case *sdl.QuitEvent: 
      c.quit = true
      break
    }
  }
}

//wait for an event and returns the value of the event in key_mapping
func (c *Chip8) waitEvent() byte{

  for {
    for c.event = sdl.PollEvent(); c.event != nil; c.event = sdl.PollEvent() {
      switch event := c.event.(type) {
      case *sdl.KeyDownEvent: 
        if code, found := c.key_mapping[event.Keysym.Sym]; found {
          return code
        } 
      }
    }
  }
}

func (chip *Chip8) cleanup() {

}


func main() {

  if len(os.Args) < 2 {
    fmt.Println("usage: emuChip8 <game>")
    os.Exit(1)
  }


  sdl.Init(sdl.INIT_EVERYTHING)
  window, err := sdl.CreateWindow("Chip8", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
      WIN_WIDTH, WIN_HEIGHT, sdl.WINDOW_SHOWN)

  if err != nil {
      panic(err)
  }
  defer window.Destroy()

  surface, err := window.GetSurface()
  if err != nil {
      panic(err)
  }

  chip := new(Chip8)
  chip.surface = surface
  chip.init()
  chip.loadGame(os.Args[1])

  chip.quit = false
  for !chip.quit {

    if err := chip.emulateCycle(); err != nil {
      fmt.Println(err)
    }

    if chip.drawFlag {
      chip.drawGraphics()
      window.UpdateSurface()
      chip.drawFlag = false
    }

    chip.setKeys()
    sdl.Delay(1)

  }

  chip.cleanup()

}

