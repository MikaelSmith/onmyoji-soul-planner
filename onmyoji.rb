#!/usr/bin/env ruby
require 'yaml'

SHIKIGAMI = {
  Onikiri: {
    ATK: 3350,
    Crit: 0.11,
    CritDMG: 1.6,
    SPD: 117
  },
  Ibaraki: {
    ATK: 3216,
    Crit: 0.10,
    CritDMG: 1.5,
    SPD: 112
  }
}

if ARGV.size < 2
  STDERR.puts "Usage: onmyoji <shikigami> <main soul> [<attr>=<constraint>]"
  exit 1
end

name, main_soul = ARGV[0], ARGV[1]
shikigami = SHIKIGAMI[name.to_sym]
if shikigami.nil?
  STDERR.puts "Unknown shikigami #{name}"
  exit 1
end

SOUL_TYPES = {
  Harpy: :ATKBonus,
  Watcher: :ATKBonus,
  'House Imp': :ATKBonus,
  Scarlet: :ATKBonus,
  Soultaker: :ATKBonus,
  Nightwing: :ATKBonus,
  Kyoukotsu: :ATKBonus,
  'Tomb Guard': :Crit,
  Shadow: :Crit,
  Fenikkusu: :Crit,
  Claws: :Crit,
  Samisen: :Crit,
  Seductress: :Crit,
  Namazu: :Variable,
  Odokuro: :Variable,
  'Ghostly Songstress': :Variable
}
unless SOUL_TYPES.include?(main_soul.to_sym)
  STDERR.puts "Unknown main soul type #{main_soul}"
  exit 1
end

constraints = ARGV[2..].each_with_object({}) do |arg, attrs|
  pair = arg.split('=')
  if pair.size != 2
    STDERR.puts "Unknown argument #{arg}, must be of the form <attribute>=<range>, such as SPD=117-127 or Crit=1.0"
    exit 1
  end
  constraint = pair[1].split('-').map(&:to_f)
  if constraint.size > 2
    STDERR.puts "Illegal constraint #{pair[1]}, must be a number N or range of the form M-N"
    exit 1
  end
  if constraint.size == 1
    constraint = [constraint.first, constraint.first]
  end
  attrs[pair[0]] = constraint
end
unless (rem = constraints.keys - %w[Crit SPD]).empty?
  STDERR.puts "Unsupported attribute constraints #{rem}"
  exit 1
end

souls_db = YAML.load_file('souls.yaml')
SLOTS = %w[Slot1 Slot2 Slot3 Slot4 Slot5 Slot6]
if (sorted = souls_db.keys.sort) != SLOTS
  STDERR.puts "Expected 6 arrays Slot1-6. Unknown keys #{sorted - SLOTS}"
  exit 1
end

ATTRIBUTES = %w[ATK ATKBonus Crit CritDMG SPD Type]
failed = false
souls_db.each do |slot, souls|
  souls.each do |soul|
    sorted = soul.keys.sort
    unless (rem = sorted - ATTRIBUTES).empty?
      STDERR.puts "Unknown attributes #{rem}"
      failed = true
    end

    unless soul['Type'] && SOUL_TYPES.include?(soul['Type'].to_sym)
      STDERR.puts "Unexpected soul type for #{soul}"
      failed = true
    end
  end
end
exit 1 if failed

def compute_crit(shiki, soul_set, types = nil)
  types ||= soul_set.each_with_object(Hash.new(0)) { |soul, ts| ts[soul['Type'].to_sym] += 1 }

  soul_crit = soul_set.reduce(0.0) { |sum, soul| sum + soul.fetch('Crit', 0.0) }
  crit = shiki[:Crit] + soul_crit
  crit_souls = 0
  SOUL_TYPES.select { |_, v| v == :Crit }.keys.each do |type|
    crit_souls += 1 if types[type] >= 2
  end
  crit += 0.15 * crit_souls
  if crit >= 1.0
    crit = 1.0
  end
  crit
end

def damage(shiki, soul_set, main_soul)
  types = soul_set.each_with_object(Hash.new(0)) { |soul, ts| ts[soul['Type'].to_sym] += 1 }
  return 0.0 if types[main_soul.to_sym] < 4

  soul_atk = soul_set.reduce(0) { |sum, soul| sum + soul.fetch('ATK', 0) }
  soul_atk_bonus = soul_set.reduce(0.0) { |sum, soul| sum + soul.fetch('ATKBonus', 0.0) }
  atk_souls = 0
  SOUL_TYPES.select { |_, v| v == :ATKBonus }.keys.each do |type|
    atk_souls += 1 if types[type] >= 2
  end
  soul_atk_bonus += 0.15 * atk_souls

  extra_atk = soul_atk + (shiki[:ATK] * soul_atk_bonus)
  atk = shiki[:ATK] + extra_atk
  #puts "Attack = #{atk}"

  crit = compute_crit(shiki, soul_set, types)
  #puts "Crit = #{crit}"

  soul_crit_dmg = soul_set.reduce(0.0) { |sum, soul| sum + soul.fetch('CritDMG', 0.0) }
  crit_dmg = shiki[:CritDMG] + soul_crit_dmg
  #puts "CritDmmg = #{crit_dmg}"

  dmg = atk * (crit * crit_dmg + (1 - crit))
  dmg *= 1.1 if types[:Odokuro] >= 2
  dmg += 1.2 * crit * atk if types[:Seductress] >= 4
  dmg
end

best_dmg = 0.0
best_souls = []
final_speed = 0
final_crit = 0.0
souls_db['Slot1'].each do |soul1|
  souls_db['Slot2'].each do |soul2|
    souls_db['Slot3'].each do |soul3|
      souls_db['Slot4'].each do |soul4|
        souls_db['Slot5'].each do |soul5|
          souls_db['Slot6'].each do |soul6|
            souls = [soul1, soul2, soul3, soul4, soul5, soul6]

            spd = souls.reduce(shikigami[:SPD]) { |sum, soul| sum + soul.fetch('SPD', 0) }
            next if (low, high = constraints['SPD']) && (spd < low || spd > high)

            crit = compute_crit(shikigami, souls)
            next if (low, high = constraints['Crit']) && (crit < low || crit > high)

            if (dmg = damage(shikigami, souls, main_soul)) > best_dmg
              best_dmg = dmg
              best_souls = souls
              final_speed = spd
              final_crit = crit
            end
          end
        end
      end
    end
  end
end

puts "#{name} dmg = #{best_dmg}, speed = #{final_speed}, crit = #{final_crit}"
puts({ 'Souls' => best_souls }.to_yaml)
