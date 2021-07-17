export default ({ cacheMap, preloadPiecesAmount, piecesInOneRow }) => {
  const cacheMapWithoutEmptyBlocks = cacheMap.filter(({ percentage }) => percentage > 0)

  const getFullAmountOfBlocks = amountOfBlocks =>
    // this function counts existed amount of blocks with extra "empty blocks" to fill the row till the end
    amountOfBlocks % piecesInOneRow === 0
      ? amountOfBlocks - 1
      : amountOfBlocks + piecesInOneRow - (amountOfBlocks % piecesInOneRow) - 1 || 0

  const amountOfBlocksToRenderInShortView = getFullAmountOfBlocks(preloadPiecesAmount)
  // preloadPiecesAmount is counted from "cache.Capacity / cache.PiecesLength". We always show at least this amount of blocks
  const scalableAmountOfBlocksToRenderInShortView = getFullAmountOfBlocks(cacheMapWithoutEmptyBlocks.length)
  // cacheMap can become bigger than preloadPiecesAmount counted before. In that case we count blocks dynamically

  const finalAmountOfBlocksToRenderInShortView = Math.max(
    // this check is needed to decide which is the biggest amount of blocks and take it to render
    scalableAmountOfBlocksToRenderInShortView,
    amountOfBlocksToRenderInShortView,
  )

  const extraBlocksAmount = finalAmountOfBlocksToRenderInShortView - cacheMapWithoutEmptyBlocks.length + 1
  // amount of blocks needed to fill the line till the end

  const extraEmptyBlocksForFillingLine = extraBlocksAmount ? new Array(extraBlocksAmount).fill({}) : []

  return [...cacheMapWithoutEmptyBlocks, ...extraEmptyBlocksForFillingLine]
}
