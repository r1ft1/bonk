<script module lang="ts">
  let installed = false
</script>

<script lang="ts">
  import type { Snippet } from 'svelte'
  import { T, useTask, useParent, useThrelte } from '@threlte/core'
  import type { CameraControlsProps } from './CameraControls.svelte'
  import { CameraControls } from './CameraControls'
  import {
    Box3,
    Matrix4,
    Quaternion,
    Raycaster,
    Sphere,
    Spherical,
    Vector2,
    Vector3,
    Vector4,
    type PerspectiveCamera
  } from 'three'
  import { DEG2RAD } from 'three/src/math/MathUtils.js'

  const subsetOfTHREE = {
    Vector2,
    Vector3,
    Vector4,
    Quaternion,
    Matrix4,
    Spherical,
    Box3,
    Sphere,
    Raycaster
  }

  if (!installed) {
    CameraControls.install({ THREE: subsetOfTHREE })
    installed = true
  }

  const parent = useParent()
  if (!$parent) {
    throw new Error('CameraControls must be a child of a ThreeJS camera')
  }

  const { renderer, invalidate } = useThrelte()

  let {
    autoRotate = false,
    autoRotateSpeed = 1,
    children,
    ...restProps
  }: CameraControlsProps & { children?: Snippet<[{ ref: CameraControls }]> } = $props()

  export const ref = new CameraControls($parent as PerspectiveCamera, renderer?.domElement)

  let disableAutoRotate = false

  useTask(
    (delta) => {
      if (autoRotate && !disableAutoRotate) {
        ref.azimuthAngle += 4 * delta * DEG2RAD * autoRotateSpeed
      }
      const updated = ref.update(delta)
      if (updated) invalidate()
    },
    { autoInvalidate: false }
  )
</script>

<T
  is={ref}
  oncontrolstart={() => { disableAutoRotate = true }}
  onzoom={(e: any) => { console.log('zoomstart', e) }}
  oncontrolend={() => { disableAutoRotate = false }}
  {...restProps}
>
  {@render children?.({ ref })}
</T>
