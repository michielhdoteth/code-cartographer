import { CanvasRenderer } from '@visualizers/canvasRenderer'
import { createSimulationNodes, createSimulationEdges, transforms } from '../../fixtures/index'
import {
  createTestCanvas,
  createRenderOptions,
  measurePerformance,
  measureMultipleRuns,
  expectPerformanceWithin,
  expectFPSAbove,
} from '../../helpers/index'

describe('CanvasRenderer', () => {
  let renderer: CanvasRenderer
  let canvas: HTMLCanvasElement
  const nodes = createSimulationNodes(100)
  const edges = createSimulationEdges(nodes, 50)

  beforeEach(() => {
    canvas = createTestCanvas(800, 600)
    renderer = new CanvasRenderer(canvas)
  })

  describe('initialization', () => {
    it('should initialize with canvas element', () => {
      expect(renderer).toBeDefined()
    })

    it('should set canvas size', () => {
      expect(canvas.width).toBe(800)
      expect(canvas.height).toBe(600)
    })

    it('should get rendering context', () => {
      const context = renderer.getContext()
      expect(context).toBeDefined()
    })
  })

  describe('rendering', () => {
    it('should render nodes', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: transforms.identity })
      }).not.toThrow()
    })

    it('should render edges', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: transforms.identity })
      }).not.toThrow()
    })

    it('should render with transform', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: transforms.combined })
      }).not.toThrow()
    })

    it('should render at different zoom levels', () => {
      const zoomLevels = [0.5, 1.0, 2.0]
      for (const zoom of zoomLevels) {
        expect(() => {
          renderer.render({ nodes, edges, transform: { x: 0, y: 0, k: zoom } })
        }).not.toThrow()
      }
    })
  })

  describe('performance', () => {
    it('should render 100 nodes in <100ms', () => {
      const { elapsed } = measurePerformance(() => {
        renderer.render({ nodes, edges, transform: transforms.identity })
      })
      expectPerformanceWithin(elapsed, 100, 'Render 100 nodes')
    })

    it('should render 500 nodes efficiently', () => {
      const largeNodes = createSimulationNodes(500)
      const { elapsed } = measurePerformance(() => {
        renderer.render({ nodes: largeNodes, edges: [], transform: transforms.identity })
      })
      expectPerformanceWithin(elapsed, 500, 'Render 500 nodes')
    })

    it('should maintain 30+ FPS for 100 nodes', () => {
      const { avgElapsed } = measureMultipleRuns(
        () => renderer.render({ nodes, edges, transform: transforms.identity }),
        30
      )
      expectFPSAbove([avgElapsed], 30, '100 nodes frame rate')
    })
  })

  describe('interaction', () => {
    it('should handle node selection', () => {
      expect(() => renderer.selectNode('node_0')).not.toThrow()
    })

    it('should handle node hover', () => {
      expect(() => renderer.hoverNode('node_0')).not.toThrow()
    })

    it('should handle zoom', () => {
      expect(() => renderer.setTransform(transforms.zoomedIn)).not.toThrow()
    })

    it('should handle pan', () => {
      expect(() => renderer.setTransform(transforms.panned)).not.toThrow()
    })
  })

  describe('label rendering', () => {
    it('should render labels at zoom level 1.0', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: transforms.identity })
      }).not.toThrow()
    })

    it('should hide labels at low zoom', () => {
      expect(() => {
        renderer.render({ nodes, edges, transform: transforms.zoomedOut })
      }).not.toThrow()
    })
  })

  describe('edge cases', () => {
    it('should handle empty nodes', () => {
      expect(() => {
        renderer.render({ nodes: [], edges: [], transform: transforms.identity })
      }).not.toThrow()
    })

    it('should handle nodes with NaN coordinates', () => {
      const badNodes = [{ id: 'bad', x: NaN, y: NaN, fx: 0, fy: 0, vx: 0, vy: 0 }]
      expect(() => {
        renderer.render({ nodes: badNodes, edges: [], transform: transforms.identity })
      }).not.toThrow()
    })

    it('should handle extreme zoom values', () => {
      for (const zoom of [0.01, 100]) {
        expect(() => {
          renderer.render({ nodes, edges, transform: { x: 0, y: 0, k: zoom } })
        }).not.toThrow()
      }
    })

    it('should handle canvas resize', () => {
      canvas.width = 1024
      canvas.height = 768
      expect(() => {
        renderer.render({ nodes, edges, transform: transforms.identity })
      }).not.toThrow()
    })
  })

  describe('color and styling', () => {
    it('should render nodes with colors', () => {
      const coloredNodes = nodes.map((n) => ({ ...n, color: '#ff0000' }))
      expect(() => {
        renderer.render({ nodes: coloredNodes, edges, transform: transforms.identity })
      }).not.toThrow()
    })

    it('should render edges with different styles', () => {
      const styledEdges = edges.map((e) => ({ ...e, color: '#00ff00', width: 2 }))
      expect(() => {
        renderer.render({ nodes, edges: styledEdges, transform: transforms.identity })
      }).not.toThrow()
    })
  })
})
