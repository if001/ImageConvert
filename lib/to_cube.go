package lib
import(
    _ "log"
    "image"
    "image/color"
    "math"
    _ "fmt"
)
type Pos struct {
    X float64 
    Y float64 
    Z float64 
}
func outImgToXYZ(i int,j int, face int, edge int) (Pos){
    var a float64 = 2.0 * float64(i)/float64(edge)
    var b float64 = 2.0 * float64(j)/float64(edge)
   
    pos := Pos{0,0,0};
    
    if(face == 0){ // back
        pos.X = -1.0
        pos.Y = 1.0-a
        pos.Z = 3.0-b
    }else if(face == 1){ //left
        pos.X = a-3.0
        pos.Y = -1.0
        pos.Z = 3.0-b
    }else if(face == 2){ //front
        pos.X = 1.0
        pos.Y = a-5.0
        pos.Z = 3.0-b
    }else if(face == 3){ //right
        pos.X = 7.0-a
        pos.Y = 1.0
        pos.Z = 3.0-b
    }else if(face == 4){ //top
        pos.X = b-1.0
        pos.Y = a-5.0
        pos.Z = 1.0
    }else if(face == 5){ //bottom
        pos.X = 5.0-b
        pos.Y = a-5.0
        pos.Z = -1.0
    }
    return pos
}

func clip(in int, min int,max int)(int){
    if(in > max){
        in = max
    }else if(in < min){
        in = min
    }
    return in
}


func createSlice(start int ,end int)( []int) {
    var s[]int
    for end > start {
        s = append(s,start)
        start += 1
    }
    return s
}

type Img struct {
    Xsize int
    Ysize int
}

func fillRect(img *image.RGBA, col color.Color) {
    rect := img.Rect
    for h := rect.Min.Y; h < rect.Max.Y; h++ {
        for v := rect.Min.X; v < rect.Max.X; v++ {
            img.Set(v, h, col)
        }
    }
}
func ToCube(in_img image.Image)(image.Image){
    var in_size = Img{
        in_img.Bounds().Dx(),
        in_img.Bounds().Dy()}

    var out_size = Img{
        in_img.Bounds().Dx(),
        int(in_img.Bounds().Dx()*3/4)}
    
    var edge = int(in_size.Xsize/4)
    var in_pix = in_img
    var new_img = image.NewRGBA(image.Rect(0, 0, out_size.Xsize,out_size.Ysize))
    fillRect(new_img, color.RGBA{0,0,0,255})
    
    for _,i := range createSlice(0, int(out_size.Xsize)){
        face := int(float64(i)/float64(edge)) // 0 - back, 1 - left 2 - front, 3 - right
        var rng[]int
        if(face == 2){
            rng = createSlice(0, int(edge*3))
        }else{
            rng = createSlice(int(edge), int(edge*2))
        }

        var face2 int
        for _,j := range rng {
            if(j < int(edge)){
                face2 = 4 // top
            }else if(j >= int(2*edge)){
                face2 = 5 // bottom
            }else{
                face2 = face
            }
            // topとbottomは除外する
            if (face2 != 4 && face2 != 5){
                pos := outImgToXYZ(i,j,face2,edge)

                var theta float64 = math.Atan2(pos.Y,pos.X)
                var r float64 = math.Hypot(pos.X, pos.Y)
                var phi float64 = math.Atan2(pos.Z,r)

                var uf float64 = 2.0*float64(edge)*(theta + math.Pi)/math.Pi
                var vf float64 = 2.0*float64(edge) * (math.Pi/2.0 - phi)/math.Pi

                var ui float64 = math.Floor(uf) // coord of pixel to bottom left
                var vi float64 = math.Floor(vf)

                var u2 float64 = ui+1 // coords of pixel to top right
                var v2 float64 = vi+1
                var mu float64 = uf-ui // fraction of way across pixel
                var nu float64 = vf-vi
                
                Ar,Ag,Ab,_ := in_pix.At(int(ui) % int(in_size.Xsize),int(clip(int(vi),0,int(in_size.Ysize-1)))).RGBA()
                Br,Bg,Bb,_ := in_pix.At(int(u2) % int(in_size.Xsize),int(clip(int(vi),0,int(in_size.Ysize-1)))).RGBA() 
                Cr,Cg,Cb,_ := in_pix.At(int(ui) % int(in_size.Xsize),int(clip(int(v2),0,int(in_size.Ysize-1)))).RGBA() 
                Dr,Dg,Db,_ := in_pix.At(int(u2) % int(in_size.Xsize),int(clip(int(v2),0,int(in_size.Ysize-1)))).RGBA() 

                red :=   float64(Ar)/257*(1-mu)*(1-nu) + float64(Br)/257*mu*(1-nu) + float64(Cr)/257*(1-mu)*nu + float64(Dr)/257*mu*nu 
                green := float64(Ag)/257*(1-mu)*(1-nu) + float64(Bg)/257*mu*(1-nu) + float64(Cg)/257*(1-mu)*nu + float64(Dg)/257*mu*nu 
                blue :=  float64(Ab)/257*(1-mu)*(1-nu) + float64(Bb)/257*mu*(1-nu) + float64(Cb)/257*(1-mu)*nu + float64(Db)/257*mu*nu

                new_img.Set(i,j,color.RGBA{uint8(red),uint8(green),uint8(blue),255})
            }
        }
    }

    return new_img
}

type Pixel struct {
    r uint8
    g uint8
    b uint8
    a uint8
}
func toPixel(r uint32,g uint32,b uint32,a uint32)(Pixel){
    return Pixel{uint8(float64(r)/257.0),uint8(float64(g)/257.0),uint8(float64(b)/257.0),uint8(float64(a)/257.0)}
}
func CutTopBottom(inp_img image.Image)(image.Image){
    width := inp_img.Bounds().Dx()
    height := inp_img.Bounds().Dy()/3
    new_img := image.NewRGBA(image.Rect(0, 0, width, height))
    for i,img_i := range createSlice(0, width) {
        for j,img_j := range createSlice(height,height*2) {   
            pixel := toPixel(inp_img.At(img_i,img_j).RGBA())
            new_img.Set(i,j,color.RGBA{pixel.r,pixel.g,pixel.b,255})
        }
    }    
    return new_img
}
