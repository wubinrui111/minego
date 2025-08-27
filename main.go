package main

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	screenWidth  = 800
	screenHeight = 600
	gravity      = 0.3  // 降低重力值从0.5到0.3
	moveSpeed    = 3
	jumpPower    = -10  // 调整跳跃力度从-12到-10
	airResistance = 0.98
)

// GameObject接口定义所有游戏对象的通用行为
type GameObject interface {
	GetX() float64
	GetY() float64
	GetWidth() float64
	GetHeight() float64
}

// 方块类型枚举（仅用于视觉区分）
const (
	PlatformType = iota
	SolidType
	WoodType
	DirtType
)

type Game struct {
	player      Player
	platforms   []Block
	camera      Camera
	score       int
	gameStarted bool
}

type Player struct {
	x, y    float64
	vx, vy  float64
	width   float64
	height  float64
	onGround bool
}

// 实现GameObject接口
func (p Player) GetX() float64 { return p.x }
func (p Player) GetY() float64 { return p.y }
func (p Player) GetWidth() float64 { return p.width }
func (p Player) GetHeight() float64 { return p.height }

type Block struct {
	x, y, width, height float64
	blockType           int
}

// 实现GameObject接口
func (b Block) GetX() float64 { return b.x }
func (b Block) GetY() float64 { return b.y }
func (b Block) GetWidth() float64 { return b.width }
func (b Block) GetHeight() float64 { return b.height }

type Camera struct {
	x, y float64
}

func (g *Game) Update() error {
	if !g.gameStarted {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.gameStarted = true
		}
		return nil
	}

	// 处理输入
	g.handleInput()

	// 应用物理
	g.applyPhysics()

	// 更新玩家位置
	g.updatePlayerPosition()

	// 检查碰撞
	g.checkCollisions()

	// 更新摄像机跟随玩家
	g.updateCamera()

	// 边界检查
	g.checkBoundaries()

	return nil
}

func (g *Game) handleInput() {
	// 水平移动
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		g.player.vx -= moveSpeed * 0.2
		if g.player.vx < -moveSpeed {
			g.player.vx = -moveSpeed
		}
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		g.player.vx += moveSpeed * 0.2
		if g.player.vx > moveSpeed {
			g.player.vx = moveSpeed
		}
	} else {
		// 添加地面摩擦力
		if g.player.onGround {
			g.player.vx *= 0.8
			if math.Abs(g.player.vx) < 0.1 {
				g.player.vx = 0
			}
		}
	}

	// 跳跃 (改为W键)
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		// 使用另一种方式实现跳跃 - 直接检查玩家下方是否有地面
		if g.canJump() {
			g.player.vy = jumpPower
			g.player.onGround = false
		}
	}
}

// canJump 检查玩家是否可以跳跃
func (g *Game) canJump() bool {
	// 如果玩家已经在地面上，则可以跳跃
	if g.player.onGround {
		return true
	}
	
	// 检查玩家下方是否有平台（另一种方式判断是否可以跳跃）
	// 创建一个位于玩家下方的临时检测区域
	tempPlayer := Player{
		x:      g.player.x,
		y:      g.player.y + 1, // 玩家下方1个像素的位置
		width:  g.player.width,
		height: g.player.height,
	}
	
	// 检查这个位置是否与任何平台相交
	for _, block := range g.platforms {
		if g.checkCollision(tempPlayer, block) {
			return true
		}
	}
	
	return false
}

func (g *Game) applyPhysics() {
	// 应用重力
	if !g.player.onGround {
		g.player.vy += gravity
	}

	// 空气阻力
	g.player.vx *= airResistance

	// 限制最大下落速度
	if g.player.vy > 10 {
		g.player.vy = 10
	}
}

func (g *Game) updatePlayerPosition() {
	// 更新玩家位置
	g.player.x += g.player.vx
	g.player.y += g.player.vy
}

func (g *Game) checkCollisions() {
	// 重置onGround状态
	g.player.onGround = false

	// 检查与所有方块的碰撞
	for _, block := range g.platforms {
		if g.checkCollision(g.player, block) {
			// 计算碰撞重叠量
			dx1 := g.player.x + g.player.width - block.x
			dx2 := block.x + block.width - g.player.x
			dy1 := g.player.y + g.player.height - block.y
			dy2 := block.y + block.height - g.player.y
			
			// 找到最小重叠方向
			minOverlap := math.Min(math.Min(dx1, dx2), math.Min(dy1, dy2))
			
			// 根据最小重叠方向解决碰撞
			if minOverlap == dx1 {
				// 从右侧撞到方块
				g.player.x = block.x - g.player.width
				g.player.vx = 0
			} else if minOverlap == dx2 {
				// 从左侧撞到方块
				g.player.x = block.x + block.width
				g.player.vx = 0
			} else if minOverlap == dy1 {
				// 从下方撞到方块（头顶撞到方块）
				g.player.y = block.y - g.player.height
				// 只有当玩家向上移动时才停止
				if g.player.vy < 0 {
					g.player.vy = 0
				}
			} else if minOverlap == dy2 {
				// 从上方撞到方块（站在方块上）
				g.player.y = block.y + block.height
				// 只有当玩家向下移动时才停止并标记为在地面上
				if g.player.vy >= 0 {
					g.player.vy = 0
					g.player.onGround = true
				}
			}
		}
	}
}

func (g *Game) updateCamera() {
	// 改进的摄像头跟随系统
	targetX := g.player.x - screenWidth/2
	targetY := g.player.y - screenHeight/2

	// 平滑跟随
	g.camera.x += (targetX - g.camera.x) * 0.1
	g.camera.y += (targetY - g.camera.y) * 0.1

	// 限制摄像机移动范围
	if g.camera.x < 0 {
		g.camera.x = 0
	}

	// 可以根据关卡大小添加更多限制
}

func (g *Game) checkBoundaries() {
	// 检查是否掉出屏幕
	if g.player.y > screenHeight + 500 {
		g.resetGame()
	}
}

func (g *Game) checkCollision(obj1, obj2 GameObject) bool {
	return obj1.GetX() < obj2.GetX()+obj2.GetWidth() &&
		obj1.GetX()+obj1.GetWidth() > obj2.GetX() &&
		obj1.GetY() < obj2.GetY()+obj2.GetHeight() &&
		obj1.GetY()+obj1.GetHeight() > obj2.GetY()
}

func (g *Game) resetGame() {
	g.player = Player{
		x:      100,
		y:      100,
		width:  20,
		height: 20,
		onGround: false, // 明确初始化onGround状态
	}
	g.score = 0
}

func (g *Game) Draw(screen *ebiten.Image) {
	// 清空屏幕
	screen.Fill(color.RGBA{50, 150, 200, 255})

	if !g.gameStarted {
		ebitenutil.DebugPrintAt(screen, "2D Gravity Game", screenWidth/2-80, screenHeight/2-60)
		ebitenutil.DebugPrintAt(screen, "Press SPACE to start", screenWidth/2-90, screenHeight/2-30)
		ebitenutil.DebugPrintAt(screen, "Controls:", screenWidth/2-40, screenHeight/2+10)
		ebitenutil.DebugPrintAt(screen, "A/D or Arrow Keys - Move", screenWidth/2-100, screenHeight/2+30)
		ebitenutil.DebugPrintAt(screen, "W - Jump", screenWidth/2-35, screenHeight/2+50)
		return
	}

	// 绘制方块
	for _, block := range g.platforms {
		switch block.blockType {
		case PlatformType:
			// 绘制木头颜色的横折线
			woodColor := color.RGBA{139, 69, 19, 255} // 棕色木头颜色
			vector.StrokeLine(screen, float32(block.x-g.camera.x), float32(block.y), 
				float32(block.x-g.camera.x+block.width), float32(block.y), 2, woodColor, false)
			vector.StrokeLine(screen, float32(block.x-g.camera.x), float32(block.y), 
				float32(block.x-g.camera.x), float32(block.y+block.height), 2, woodColor, false)
			vector.StrokeLine(screen, float32(block.x-g.camera.x+block.width), float32(block.y), 
				float32(block.x-g.camera.x+block.width), float32(block.y+block.height), 2, woodColor, false)
			vector.StrokeLine(screen, float32(block.x-g.camera.x), float32(block.y+block.height), 
				float32(block.x-g.camera.x+block.width), float32(block.y+block.height), 2, woodColor, false)
		case SolidType:
			// 绘制红色固体方块
			ebitenutil.DrawRect(screen, block.x-g.camera.x, block.y, block.width, block.height, color.RGBA{255, 0, 0, 255})
		case WoodType:
			// 绘制木块
			ebitenutil.DrawRect(screen, block.x-g.camera.x, block.y, block.width, block.height, color.RGBA{139, 69, 19, 255})
		case DirtType:
			// 绘制棕色土方块
			ebitenutil.DrawRect(screen, block.x-g.camera.x, block.y, block.width, block.height, color.RGBA{150, 75, 0, 255})
		}
	}

	// 绘制玩家（始终是蓝色的方块）
	playerColor := color.RGBA{50, 50, 255, 255} // 蓝色
	ebitenutil.DrawRect(screen, g.player.x-g.camera.x, g.player.y, g.player.width, g.player.height, playerColor)

	// 绘制分数
	ebitenutil.DebugPrint(screen, fmt.Sprintf("Score: %d", g.score))
	ebitenutil.DebugPrintAt(screen, "Controls: A/D - Move, W - Jump", 10, 10)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

// Jump 实现玩家跳跃功能
func (g *Game) Jump() {
	// 检查玩家是否可以跳跃（在地面上）
	if g.player.onGround {
		// 给玩家一个向上的速度
		g.player.vy = jumpPower
		// 玩家离开地面
		g.player.onGround = false
	}
}

func NewGame() *Game {
	game := &Game{
		player: Player{
			x:      100,
			y:      100,
			width:  20,
			height: 20,
			onGround: false, // 明确初始化onGround状态
		},
		platforms: []Block{
			// 地面
			{x: 0, y: screenHeight - 40, width: screenWidth * 2, height: 40, blockType: DirtType},
			// 平台 (木头颜色的横折线)
			{x: 300, y: screenHeight - 120, width: 200, height: 20, blockType: PlatformType},
			{x: 600, y: screenHeight - 200, width: 150, height: 20, blockType: PlatformType},
			{x: 200, y: screenHeight - 300, width: 100, height: 20, blockType: PlatformType},
			{x: 500, y: screenHeight - 380, width: 250, height: 20, blockType: PlatformType},
			// 红色固体方块 (现在也具有相同性质)
			{x: 400, y: screenHeight - 250, width: 30, height: 100, blockType: SolidType},
			// 棕色土方块
			{x: 100, y: screenHeight - 100, width: 50, height: 60, blockType: DirtType},
			// 木块
			{x: 700, y: screenHeight - 300, width: 40, height: 40, blockType: WoodType},
		},
	}

	return game
}

func main() {
	game := NewGame()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("2D Gravity Game")
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}