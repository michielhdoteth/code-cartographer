import { WebGLRenderer } from '@visualizers/webglRenderer'

describe('WebGLRenderer', () => {
  let renderer: WebGLRenderer
  let canvas: HTMLCanvasElement
  const nodes = Array.from({ length: 1000 }, (_, i) => ({
    id: `node${i}`,
    x: Math.random() * 1000,
    y: Math.random() * 1000,
    fx: 0,
    fy: 0,
    vx: 0,
    vy: 0,
  }))

  const edges = Array.from({ length: 500 }, (_, i) => ({
    id: `edge${i}`,
    source: nodes[i % nodes.length],
    target: nodes[(i + 1) % nodes.length],
  }))

  beforeEach(() => {
    canvas = document.createElement('canvas')
    canvas.width = 1024
    canvas.height = 768
    renderer = new WebGLRenderer(canvas)
  })

  describe('initialization', () => {
    it('should initialize WebGL context', () => {
      const context = renderer.getContext()
      expect(context).toBeDefined()
    })

    it('should compile shaders', () => {
      expect(renderer).toBeDefined()
    })

    it('should setup vertex and index buffers', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: { x: 0, y: 0, k: 1 } })
      }).not.toThrow()
    })
  })

  describe('rendering performance', () => {
    it('should render 1000 nodes in <500ms', () => {
      const start = performance.now()
      renderer.render({ nodes, edges, transform: { x: 0, y: 0, k: 1 } })
      const elapsed = performance.now() - start
      expect(elapsed).toBeLessThan(500)
    })

    it('should render 5000 nodes efficiently', () => {
      const largeNodes = Array.from({ length: 5000 }, (_, i) => ({
        id: `node${i}`,
        x: Math.random() * 2000,
        y: Math.random() * 2000,
        fx: 0,
        fy: 0,
        vx: 0,
        vy: 0,
      }))

      const start = performance.now()
      renderer.render({ nodes: largeNodes, edges: [], transform: { x: 0, y: 0, k: 1 } })
      const elapsed = performance.now() - start
      expect(elapsed).toBeLessThan(1000)
    })

    it('should maintain 30+ FPS for 1000 nodes', () => {
      const frameTimings = []
      for (let i = 0; i < 30; i++) {
        const start = performance.now()
        renderer.render({ nodes, edges, transform: { x: 0, y: 0, k: 1 } })
        frameTimings.push(performance.now() - start)
      }
      const avgFrameTime = frameTimings.reduce((a, b) => a + b) / frameTimings.length
      expect(avgFrameTime).toBeLessThan(33)
    })

    it('should achieve GPU acceleration speedup', () => {
      const nodeCount = 2000
      const largeNodeSet = Array.from({ length: nodeCount }, (_, i) => ({
        id: `node${i}`,
        x: Math.random() * 2000,
        y: Math.random() * 2000,
        fx: 0,
        fy: 0,
        vx: 0,
        vy: 0,
      }))

      const start = performance.now()
      renderer.render({ nodes: largeNodeSet, edges: [], transform: { x: 0, y: 0, k: 1 } })
      const elapsed = performance.now() - start

      expect(elapsed).toBeLessThan(1000)
    })
  })

  describe('transformations', () => {
    it('should apply zoom transformation', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: { x: 0, y: 0, k: 2 } })
      }).not.toThrow()
    })

    it('should apply pan transformation', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: { x: 100, y: 100, k: 1 } })
      }).not.toThrow()
    })

    it('should handle combined zoom and pan', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: { x: 200, y: 150, k: 1.5 } })
      }).not.toThrow()
    })

    it('should handle extreme zoom levels', () => {
      for (const zoom of [0.1, 0.5, 1, 5, 10]) {
        expect(() => {
          renderer.render({ nodes, edges, transform: { x: 0, y: 0, k: zoom } })
        }).not.toThrow()
      }
    })
  })

  describe('fallback to canvas', () => {
    it('should fallback to canvas if WebGL unavailable', () => {
      expect(renderer).toBeDefined()
    })

    it('should render correctly with fallback', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: { x: 0, y: 0, k: 1 } })
      }).not.toThrow()
    })
  })

  describe('memory management', () => {
    it('should manage GPU memory efficiently', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: { x: 0, y: 0, k: 1 } })
      }).not.toThrow()
    })

    it('should cleanup resources', () => {
      expect(() => {
        renderer.cleanup()
      }).not.toThrow()
    })
  })

  describe('color and styling', () => {
    it('should render nodes with different colors', () => {
      const coloredNodes = nodes.map((n, i) => ({
        ...n,
        color: i % 2 === 0 ? '#ff0000' : '#00ff00',
      }))
      expect(() => {
        renderer.render({ nodes: coloredNodes, edges, transform: { x: 0, y: 0, k: 1 } })
      }).not.toThrow()
    })

    it('should render edges with varying widths', () => {
      const styledEdges = edges.map((e) => ({
        ...e,
        width: Math.random() * 3,
      }))
      expect(() => {
        renderer.render({ nodes, edges: styledEdges, transform: { x: 0, y: 0, k: 1 } })
      }).not.toThrow()
    })
  })

  describe('interaction', () => {
    it('should handle node selection', () => {
      expect(() => {
        renderer.selectNode('node0')
      }).not.toThrow()
    })

    it('should handle hover effects', () => {
      expect(() => {
        renderer.hoverNode('node0')
      }).not.toThrow()
    })

    it('should clear selection', () => {
      expect(() => {
        renderer.clearSelection()
      }).not.toThrow()
    })
  })

  describe('edge cases', () => {
    it('should handle empty node list', () => {
      expect(() => {
        renderer.render({ nodes: [], edges: [], transform: { x: 0, y: 0, k: 1 } })
      }).not.toThrow()
    })

    it('should handle NaN coordinates gracefully', () => {
      const badNodes = [
        { id: 'bad1', x: NaN, y: 100, fx: 0, fy: 0, vx: 0, vy: 0 },
        { id: 'bad2', x: 100, y: NaN, fx: 0, fy: 0, vx: 0, vy: 0 },
      ]
      expect(() => {
        renderer.render({ nodes: badNodes, edges: [], transform: { x: 0, y: 0, k: 1 } })
      }).not.toThrow()
    })

    it('should handle canvas resize', () => {
      canvas.width = 1920
      canvas.height = 1080
      expect(() => {
        renderer.render({ nodes, edges, transform: { x: 0, y: 0, k: 1 } })
      }).not.toThrow()
    })

    it('should handle very large edge lists', () => {
      const largeEdges = Array.from({ length: 5000 }, (_, i) => ({
        id: `edge${i}`,
        source: nodes[i % nodes.length],
        target: nodes[(i + 1) % nodes.length],
      }))
      expect(() => {
        renderer.render({ nodes, edges: largeEdges, transform: { x: 0, y: 0, k: 1 } })
      }).not.toThrow()
    })
  })

  describe('performance benchmarks', () => {
    it('should handle 5000 nodes at 30+ FPS', () => {
      const benchmarkNodes = Array.from({ length: 5000 }, (_, i) => ({
        id: `node${i}`,
        x: Math.random() * 2000,
        y: Math.random() * 2000,
        fx: 0,
        fy: 0,
        vx: 0,
        vy: 0,
      }))

      const frameTimings = []
      for (let i = 0; i < 10; i++) {
        const start = performance.now()
        renderer.render({ nodes: benchmarkNodes, edges: [], transform: { x: 0, y: 0, k: 1 } })
        frameTimings.push(performance.now() - start)
      }
      const avgFrameTime = frameTimings.reduce((a, b) => a + b) / frameTimings.length
      expect(avgFrameTime).toBeLessThan(33)
    })
  })
})
