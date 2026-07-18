import * as v from 'valibot'

export const CAPABILITY_GROUPS = {
  control: ['control.delay', 'control.millis'],
  io: ['io.pinMode', 'io.digitalWrite', 'io.digitalRead', 'io.analogRead', 'io.pwm', 'io.tone'],
  serial: ['serial.begin', 'serial.print', 'serial.println', 'serial.available', 'serial.read'],
  motor: ['motor.servo', 'motor.stepper'],
  sensor: [
    'sensor.button', 'sensor.ldr', 'sensor.pir', 'sensor.joystick',
    'sensor.ultrasonic', 'sensor.ping',
    'sensor.tmp', 'sensor.dht', 'sensor.ds18b20', 'sensor.ntc', 'sensor.bh1750',
    'sensor.soil', 'sensor.waterlevel', 'sensor.rain', 'sensor.mq135', 'sensor.mq7',
    'sensor.sound', 'sensor.bmp', 'sensor.mlx90614', 'sensor.pulse'
  ],
  electricity: [
    'electricity.acs712', 'electricity.voltage', 'electricity.ina219',
    'electricity.zmpt101b', 'electricity.power', 'electricity.energy'
  ],
  keypad: ['keypad.init', 'keypad.read'],
  screen: ['screen.lcd', 'screen.oled'],
  security: ['security.fire', 'security.door', 'security.rfid', 'security.ir', 'security.vibration'],
  radio: ['radio.wifi', 'radio.bluetooth']
} as const

export const ALL_CAPABILITIES = [
  ...CAPABILITY_GROUPS.control,
  ...CAPABILITY_GROUPS.io,
  ...CAPABILITY_GROUPS.serial,
  ...CAPABILITY_GROUPS.motor,
  ...CAPABILITY_GROUPS.sensor,
  ...CAPABILITY_GROUPS.electricity,
  ...CAPABILITY_GROUPS.keypad,
  ...CAPABILITY_GROUPS.screen,
  ...CAPABILITY_GROUPS.security,
  ...CAPABILITY_GROUPS.radio
] as const

export const CapabilitySchema = v.picklist(ALL_CAPABILITIES)
export type Capability = v.InferOutput<typeof CapabilitySchema>

export const BoardSchema = v.object({
  id: v.pipe(v.string(), v.regex(/^[a-z0-9-]+$/, 'id debe ser kebab-case')),
  name: v.pipe(v.string(), v.nonEmpty()),
  fqbn: v.string(),
  core: v.nullable(v.string()),
  capabilities: v.array(CapabilitySchema),
  pins: v.object({
    digital: v.array(v.string()),
    analog: v.array(v.string()),
    pwm: v.array(v.string())
  }),
  defaultBaud: v.pipe(v.number(), v.integer()),
  illustration: v.optional(v.string())
})

export type Board = v.InferOutput<typeof BoardSchema>

export const BOARD_NONE: Board = {
  id: 'none',
  name: 'Sin placa',
  fqbn: '',
  core: null,
  capabilities: [],
  pins: { digital: [], analog: [], pwm: [] },
  defaultBaud: 9600
}
