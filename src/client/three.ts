/// <reference lib="dom" />
import * as THREE from 'three'


// 1. Rename the class and geometry/mesh properties
class TetrahedronScene {
  containerId: string
  container: HTMLElement | null
  scene: THREE.Scene
  // 2. Change the geometry type to TetrahedronGeometry
  tetrahedronGeometry: THREE.TetrahedronGeometry
  torusMaterial: THREE.MeshPhongMaterial // Material can stay the same name
  torus: THREE.Mesh // Mesh can stay the same name
  renderer: THREE.WebGLRenderer | undefined = undefined
  camera: THREE.PerspectiveCamera | undefined = undefined
  spotlight: THREE.DirectionalLight | undefined = undefined
  lightX: number = 1
  lightY: number = 1
  lightZ: number = 1
  ambientLight: THREE.AmbientLight = new THREE.AmbientLight(0xffffff)
  
  // 3. Simplify the constructor arguments to just 'radius' and 'detail' (optional)
  constructor(containerId: string, radius: number, detail: number = 0) {
    this.containerId = containerId
    this.scene = new THREE.Scene()
    
    // 4. Create the new geometry: TetrahedronGeometry(radius, detail)
    this.tetrahedronGeometry = new THREE.TetrahedronGeometry(radius, detail)
    
    this.torusMaterial = new THREE.MeshPhongMaterial({
      color: 0xf802fa, // Changed color to green for distinction
      shininess: 30
    })
    
    // Use the new geometry
    this.torus = new THREE.Mesh(this.tetrahedronGeometry, this.torusMaterial)
    
    this.container = document.querySelector(containerId)
    if (!this.container) {
      console.error(`element with id ${containerId} not found`)
      return
    }
    this.camera = new THREE.PerspectiveCamera(75, this.container.clientWidth / this.container.clientHeight, 0.1, 1000)
    
    // Corrected the typo in the DirectionalLight color from 0xfffff to 0xffffff
    this.renderer = new THREE.WebGLRenderer()
    this.spotlight = new THREE.DirectionalLight(0xffffff, 3) 
    
    this.spotlight.position.set(this.lightX, this.lightY, this.lightZ)
    this.scene.add(this.spotlight)
    this.scene.add(this.ambientLight)
    this.scene.add(this.torus)
    this.scene.background = new THREE.Color(0xffffff)
  }
  
  public initAndRender() {
    if (!this.container || !this.renderer || !this.camera) {
      console.error('Initialization failed. Cannot render.')
      return
    }
    this.renderer.setSize(this.container.clientWidth, this.container.clientHeight)
    this.container.appendChild(this.renderer.domElement)
    
    // 5. Adjust Camera Position: A radius of 20 used below is large for a Tetrahedron, 
    // so we pull the camera back further to see the whole shape.
    this.camera.position.z = 40; 
    
    this.renderer.setAnimationLoop(this.animate);
    window.addEventListener('resize', this.resize);
  }
  // ... resize and animate methods remain the same ...
  public resize = () => {
    if (this.container && this.camera && this.renderer) {
      const newWidth = this.container.clientWidth;
      const newHeight = this.container.clientHeight;
      this.camera.aspect = newWidth / newHeight;
      this.camera.updateProjectionMatrix();
      this.renderer.setSize(newWidth, newHeight);
    }
  }

  public animate = () => {
    this.torus.rotation.x += 0.01;
    this.torus.rotation.y += 0.02;
    if (this.renderer && this.camera) {
      this.renderer.render(this.scene, this.camera);
    }
  }
}

// 6. Instantiate the new class with appropriate parameters (radius 20, detail 0)
let tetrahedronScene = new TetrahedronScene('#three-screen', 20, 0)
tetrahedronScene.initAndRender()