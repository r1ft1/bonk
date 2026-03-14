import type { Props } from '@threlte/core'
import CC from 'camera-controls'
import type { Component } from 'svelte'
export type CameraControlsProps = Props<CC> & {
  autoRotate?: boolean
  autoRotateSpeed?: number
}
declare const CameraControls: Component<CameraControlsProps>
export default CameraControls
