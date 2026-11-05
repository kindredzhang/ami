import { describe, expect, it } from 'vitest'
import { macWindowChromeOptions } from './window-chrome'

describe('macWindowChromeOptions', () => {
  it('uses hidden title-bar chrome on macOS', () => {
    expect(macWindowChromeOptions('darwin', 'ami-chat')).toMatchObject({
      tabbingIdentifier: 'ami-chat',
      titleBarStyle: 'hidden',
      trafficLightPosition: { x: 14, y: 13 },
    })
  })

  it('leaves other platforms on their standard opaque window chrome', () => {
    expect(macWindowChromeOptions('win32', 'ami-chat')).toEqual({})
    expect(macWindowChromeOptions('linux', 'ami-chat')).toEqual({})
  })
})
