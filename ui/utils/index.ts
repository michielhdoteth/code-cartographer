export { hashContent, hashContentSync, generateCacheKey, generateCacheKeySync, hasContentLikelyChanged } from './hashUtils.js'
export { MemoizationCache, AsyncMemoizationCache, memoize } from './memoization.js'
export type { MemoizeOptions } from './memoization.js'
export {
  sanitizeFilePath,
  sanitizeGitCommand,
  validateLanguage,
  isGitRepository,
  validateFileSize,
  validateFileExtension,
  sanitizeAuthorName,
  validatePort,
  validateUrl,
  containsSuspiciousPatterns,
  validateJson,
} from './validation.js'
export {
  TYPE_COLORS,
  LAYER_COLORS,
  LANGUAGE_COLORS,
  FOLDER_COLORS,
  DEFAULT_COLOR,
  stringHash,
  getFolderColor,
  getTypeColor,
  getComplexityColor,
  getChurnColor,
  calculateBounds,
  calculateUniformScale,
  calculateCentroid,
  boundsIntersect,
  distance,
  distanceSquared,
  clamp,
  lerp,
  worldToScreen,
  screenToWorld,
} from './visualizationUtils.js'
export type { Bounds, RectBounds, Point2D } from './visualizationUtils.js'
