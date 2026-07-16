export default () => {
  if (typeof window === `undefined` || typeof navigator === `undefined`) return false

  return /iPhone|iPad|iPod/i.test(navigator.userAgent || navigator.vendor)
}
