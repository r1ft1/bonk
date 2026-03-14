import type { Events, Props, Slots } from '@threlte/core'
import CC from 'camera-controls'
import type { Component } from 'svelte'
export type CameraControlsProps = Props<CC> & {
  autoRotate?: boolean
  autoRotateSpeed?: number
}
export type CameraControlsEvents = Events<CC>
export type CameraControlsSlots = Slots<CC>
declare const CameraControls: Component<CameraControlsProps>
export default CameraControls
