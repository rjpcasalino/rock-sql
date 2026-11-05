-- Astronomy schema and seed data

DROP TABLE IF EXISTS moons CASCADE;
DROP TABLE IF EXISTS planets CASCADE;
DROP TABLE IF EXISTS exoplanets CASCADE;
DROP TABLE IF EXISTS messier_catalog CASCADE;
DROP TABLE IF EXISTS constellations CASCADE;

-- 1. Planets & Dwarf Planets
CREATE TABLE planets (
    name text PRIMARY KEY,
    type text NOT NULL,
    order_from_sun integer NOT NULL,
    distance_from_sun_au numeric(6,3) NOT NULL,
    diameter_km integer NOT NULL,
    mass_earth_relative numeric(10,4) NOT NULL,
    gravity_m_s2 numeric(6,2) NOT NULL,
    orbital_period_days numeric(10,2) NOT NULL,
    rotation_period_hours numeric(10,2) NOT NULL,
    moons_count integer NOT NULL DEFAULT 0,
    has_rings boolean NOT NULL DEFAULT false,
    mean_temp_c integer,
    surface_pressure_bar numeric(10,4),
    atmosphere_primary text,
    discovery_year integer,
    discoverer text,
    color_hex text NOT NULL,
    description text NOT NULL
);

INSERT INTO planets VALUES
('Mercury', 'Terrestrial Planet', 1, 0.387, 4879, 0.0553, 3.70, 87.97, 1407.60, 0, false, 167, 0, 'Trace (O2, Na, H2)', NULL, 'Known to ancients', '#9e9e9e', 'Smallest planet and closest to the Sun; experiences extreme temperature swings with heavily cratered terrain.'),
('Venus', 'Terrestrial Planet', 2, 0.723, 12104, 0.8150, 8.87, 224.70, -5832.50, 0, false, 464, 92.0000, '96.5% CO2, 3.5% N2', NULL, 'Known to ancients', '#e3bb76', 'Hottest planet in the Solar System, trapped beneath crushing carbon dioxide clouds under a runaway greenhouse effect.'),
('Earth', 'Terrestrial Planet', 3, 1.000, 12742, 1.0000, 9.81, 365.26, 23.93, 1, false, 15, 1.0130, '78.1% N2, 20.9% O2, 0.9% Ar', NULL, 'Known to ancients', '#6b93d6', 'Our home planet, the only known world harboring surface liquid water oceans and flourishing life.'),
('Mars', 'Terrestrial Planet', 4, 1.524, 6779, 0.1070, 3.72, 686.98, 24.62, 2, false, -65, 0.0060, '95.3% CO2, 2.6% N2, 1.9% Ar', NULL, 'Known to ancients', '#c1440e', 'The Red Planet, featuring Olympus Mons (tallest volcano in Solar System) and the immense Valles Marineris canyon.'),
('Ceres', 'Dwarf Planet', 5, 2.767, 939, 0.00016, 0.28, 1682.00, 9.07, 0, false, -106, 0, 'Trace water vapor exosphere', 1801, 'Giuseppe Piazzi', '#7a7a7a', 'Largest body in the main asteroid belt between Mars and Jupiter; contains a rocky core and an icy mantle.'),
('Jupiter', 'Gas Giant', 6, 5.204, 139820, 317.8300, 24.79, 4332.59, 9.93, 95, true, -110, NULL, '89.8% H2, 10.2% He', NULL, 'Known to ancients', '#d8ca9d', 'The giant of the Solar System with more mass than all other planets combined, famous for the colossal Great Red Spot storm.'),
('Saturn', 'Gas Giant', 7, 9.582, 116460, 95.1600, 10.44, 10759.22, 10.57, 146, true, -140, NULL, '96.3% H2, 3.25% He', NULL, 'Known to ancients', '#ead6b8', 'Famed for its dazzling, extensive ring system of orbiting water-ice particles, rock fragments, and dust.'),
('Uranus', 'Ice Giant', 8, 19.201, 50724, 14.5400, 8.69, 30685.40, -17.24, 28, true, -195, NULL, '82.5% H2, 15.2% He, 2.3% CH4', 1781, 'William Herschel', '#b2d8d8', 'An ice giant rotating on its side at an extreme 98-degree axial tilt, resulting in dramatic 21-year seasonal cycles.'),
('Neptune', 'Ice Giant', 9, 30.047, 49244, 17.1500, 11.15, 60189.00, 16.11, 16, true, -200, NULL, '80% H2, 19% He, 1.5% CH4', 1846, 'Urbain Le Verrier & Johann Galle', '#5b5ddf', 'Outermost major planet, swept by supersonic winds exceeding 2,100 km/h with deep methane-rich azure blue coloration.'),
('Pluto', 'Dwarf Planet', 10, 39.482, 2376, 0.0022, 0.62, 90560.00, -153.29, 5, false, -229, 0.00001, '98% N2, methane, CO', 1930, 'Clyde Tombaugh', '#c4a482', 'Kuiper belt dwarf planet showcasing the nitrogen-ice glacier plain Tombaugh Regio (the celestial heart).'),
('Haumea', 'Dwarf Planet', 11, 43.130, 1595, 0.0007, 0.40, 103774.00, 3.92, 2, true, -241, 0, 'None detected', 2004, 'Sierra Nevada / Palomar', '#a3a3a3', 'Rapidly rotating Kuiper belt dwarf planet distorted into an elongated triaxial ellipsoid, surrounded by its own ring.'),
('Makemake', 'Dwarf Planet', 12, 45.790, 1430, 0.0005, 0.50, 112897.00, 22.83, 1, false, -239, 0, 'Trace methane / ethane', 2005, 'Michael E. Brown et al.', '#b88968', 'Bright reddish dwarf planet in the Kuiper belt encrusted with frozen methane and ethane frost.'),
('Eris', 'Dwarf Planet', 13, 67.780, 2326, 0.0028, 0.82, 203830.00, 25.90, 1, false, -231, 0, 'Frozen methane surface', 2005, 'Michael E. Brown et al.', '#e0e0e0', 'The most massive known dwarf planet in the Solar System, residing in the distant scattered disc beyond the Kuiper belt.');

-- 2. Major Moons
CREATE TABLE moons (
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    planet_name text NOT NULL REFERENCES planets(name),
    diameter_km numeric(8,1) NOT NULL,
    mass_kg numeric(28,4),
    orbital_period_days numeric(10,4) NOT NULL,
    distance_from_planet_km integer NOT NULL,
    has_subsurface_ocean boolean NOT NULL DEFAULT false,
    has_atmosphere boolean NOT NULL DEFAULT false,
    discovery_year integer,
    discoverer text,
    notes text
);

INSERT INTO moons (name, planet_name, diameter_km, mass_kg, orbital_period_days, distance_from_planet_km, has_subsurface_ocean, has_atmosphere, discovery_year, discoverer, notes) VALUES
('Moon', 'Earth', 3474.8, 7.342e22, 27.3217, 384400, false, false, NULL, 'Known to ancients', 'Only natural satellite of Earth; stabilizes our axial tilt and drives oceanic tides.'),
('Phobos', 'Mars', 22.2, 1.066e16, 0.3189, 9376, false, false, 1877, 'Asaph Hall', 'Orbits closer to its primary than any other moon in the Solar System; gradually spiraling inward.'),
('Deimos', 'Mars', 12.4, 1.476e15, 1.2630, 23463, false, false, 1877, 'Asaph Hall', 'Outermost, smaller Martian moon covered in a thick blanket of regolith dust.'),
('Io', 'Jupiter', 3643.2, 8.932e22, 1.7691, 421700, false, true, 1610, 'Galileo Galilei', 'Most volcanically active body in the Solar System with over 400 active silicate volcanoes.'),
('Europa', 'Jupiter', 3121.6, 4.800e22, 3.5512, 670900, true, true, 1610, 'Galileo Galilei', 'Cracked ice shell concealing a global liquid water ocean harboring more volume than all Earth oceans combined.'),
('Ganymede', 'Jupiter', 5268.2, 1.482e23, 7.1546, 1070400, true, true, 1610, 'Galileo Galilei', 'Largest moon in the Solar System (larger than Mercury); possesses its own internally generated magnetic field.'),
('Callisto', 'Jupiter', 4820.6, 1.076e23, 16.6890, 1882700, true, false, 1610, 'Galileo Galilei', 'Heavily cratered, geologically inert ancient world with possible deep subsurface saline ocean.'),
('Amalthea', 'Jupiter', 167.0, 2.080e18, 0.4982, 181400, false, false, 1892, 'Edward Emerson Barnard', 'Irregular inner moon; emits more heat than it receives from the Sun due to Jovian radiation currents.'),
('Himalia', 'Jupiter', 139.6, 4.200e18, 250.5600, 11461000, false, false, 1904, 'Charles Dillon Perrine', 'Largest irregular Jovian satellite; prominent member of the Himalia group.'),
('Mimas', 'Saturn', 396.4, 3.750e19, 0.9424, 185520, true, false, 1789, 'William Herschel', 'Dominated by the Herschel impact crater, giving it a resemblance to the Death Star; harbors young ocean.'),
('Enceladus', 'Saturn', 504.2, 1.080e20, 1.3702, 238020, true, true, 1789, 'William Herschel', 'Cryovolcanic geysers erupting water vapor, silica, and organics from a global subsurface ocean into space.'),
('Tethys', 'Saturn', 1062.0, 6.174e20, 1.8878, 294660, false, false, 1684, 'Giovanni Domenico Cassini', 'Composed almost entirely of pure water ice; marked by the giant Odysseus crater and Ithaca Chasma.'),
('Dione', 'Saturn', 1122.8, 1.095e21, 2.7369, 377400, true, true, 1684, 'Giovanni Domenico Cassini', 'Dense icy moon with bright ice cliffs on its trailing hemisphere and evidence of a subsurface ocean.'),
('Rhea', 'Saturn', 1527.6, 2.307e21, 4.5175, 527040, false, true, 1672, 'Giovanni Domenico Cassini', 'Second largest moon of Saturn; tenuous exosphere of oxygen and carbon dioxide sustained by radiation.'),
('Titan', 'Saturn', 5149.5, 1.345e23, 15.9454, 1221870, true, true, 1655, 'Christiaan Huygens', 'Dense nitrogen-methane atmosphere, liquid hydrocarbon lakes and seas, river deltas, and interior ocean.'),
('Hyperion', 'Saturn', 270.0, 5.620e18, 21.2766, 1481100, false, false, 1848, 'W.C. Bond & W. Lassell', 'Spongy, deeply pitted chaotic rotator with exceptionally high porosity.'),
('Iapetus', 'Saturn', 1469.0, 1.806e21, 79.3302, 3561300, false, false, 1671, 'Giovanni Domenico Cassini', 'Striking two-tone albedo world (coal-black leading side, bright ice trailing side) with an equatorial ridge.'),
('Phoebe', 'Saturn', 213.0, 8.292e18, 550.5600, 12947780, false, false, 1899, 'William Henry Pickering', 'Retrograde orbit suggesting capture from the Kuiper belt; source of Saturn''s massive Phoebe dust ring.'),
('Miranda', 'Uranus', 471.6, 6.590e19, 1.4135, 129390, true, false, 1948, 'Gerard Kuiper', 'Dramatic patchwork surface boasting Verona Rupes, the tallest sheer cliff in the Solar System (~20 km high).'),
('Ariel', 'Uranus', 1157.8, 1.353e21, 2.5204, 191020, true, false, 1851, 'William Lassell', 'Brightest and geologically youngest surface among Uranus satellites, carved by vast graben canyons.'),
('Umbriel', 'Uranus', 1169.4, 1.172e21, 4.1442, 266300, false, false, 1851, 'William Lassell', 'Darkest of the Uranian major moons, covered in ancient impact craters.'),
('Titania', 'Uranus', 1577.8, 3.527e21, 8.7062, 435910, true, false, 1787, 'William Herschel', 'Largest satellite of Uranus; laced with enormous fault scarps and likely maintaining a liquid mantle layer.'),
('Oberon', 'Uranus', 1522.8, 3.014e21, 13.4632, 583520, false, false, 1787, 'William Herschel', 'Outermost major moon of Uranus; heavily cratered with mysterious dark carbonaceous deposits on crater floors.'),
('Triton', 'Neptune', 2706.8, 2.140e22, -5.8770, 354760, true, true, 1846, 'William Lassell', 'Only large moon in the Solar System with a retrograde orbit; active nitrogen geysers and cantaloupe terrain.'),
('Proteus', 'Neptune', 420.0, 4.400e19, 1.1223, 117647, false, false, 1989, 'Voyager 2', 'Largest irregular, non-spherical moon of Neptune; maximum size a body can achieve without self-gravity rounding.'),
('Nereid', 'Neptune', 340.0, 3.100e19, 360.1362, 5513400, false, false, 1949, 'Gerard Kuiper', 'Unusually eccentric orbit ranging from 1.4 to 9.6 million kilometers from Neptune.'),
('Charon', 'Pluto', 1212.0, 1.586e21, 6.3872, 19591, false, false, 1978, 'James Christy', 'Mutually tidally locked binary system with Pluto; reddish tholin-stained north polar region (Mordor Macula).'),
('Nix', 'Pluto', 49.8, 4.500e16, 24.8550, 48694, false, false, 2005, 'Hubble Space Telescope', 'Elongated, chaotic tumbling moon reflecting light off pure water ice.'),
('Hydra', 'Pluto', 50.9, 4.800e16, 38.2020, 64738, false, false, 2005, 'Hubble Space Telescope', 'Outermost known moon of Pluto; spins rapidly about 89 times per orbit.');

-- 3. Notable Confirmed Exoplanets
CREATE TABLE exoplanets (
    id serial PRIMARY KEY,
    name text NOT NULL UNIQUE,
    host_star text NOT NULL,
    constellation text,
    distance_light_years numeric(8,2) NOT NULL,
    discovery_method text NOT NULL,
    discovery_year integer NOT NULL,
    mass_earth_relative numeric(10,3),
    radius_earth_relative numeric(10,3),
    orbital_period_days numeric(12,4) NOT NULL,
    habitable_zone boolean NOT NULL DEFAULT false,
    planet_type text NOT NULL,
    equilibrium_temp_k integer
);

INSERT INTO exoplanets (name, host_star, constellation, distance_light_years, discovery_method, discovery_year, mass_earth_relative, radius_earth_relative, orbital_period_days, habitable_zone, planet_type, equilibrium_temp_k) VALUES
('Proxima Centauri b', 'Proxima Centauri', 'Centaurus', 4.25, 'Radial Velocity', 2016, 1.173, 1.030, 11.1860, true, 'Terrestrial', 234),
('Barnard''s Star b', 'Barnard''s Star', 'Ophiuchus', 5.96, 'Radial Velocity', 2024, 0.370, 0.750, 3.1533, false, 'Terrestrial', 400),
('Ross 128 b', 'Ross 128', 'Virgo', 11.03, 'Radial Velocity', 2017, 1.400, 1.110, 9.8658, true, 'Terrestrial', 280),
('Gliese 667 Cc', 'Gliese 667 C', 'Scorpius', 23.62, 'Radial Velocity', 2011, 3.800, 1.540, 28.1400, true, 'Super-Earth', 277),
('TRAPPIST-1e', 'TRAPPIST-1', 'Aquarius', 40.66, 'Transit', 2017, 0.692, 0.920, 6.0996, true, 'Terrestrial', 251),
('TRAPPIST-1d', 'TRAPPIST-1', 'Aquarius', 40.66, 'Transit', 2017, 0.388, 0.788, 4.0496, true, 'Terrestrial', 288),
('TRAPPIST-1f', 'TRAPPIST-1', 'Aquarius', 40.66, 'Transit', 2017, 1.039, 1.045, 9.2067, true, 'Terrestrial', 219),
('55 Cancri e', '55 Cancri', 'Cancer', 41.06, 'Radial Velocity', 2004, 8.630, 1.875, 0.7365, false, 'Super-Earth', 2200),
('LHS 1140 b', 'LHS 1140', 'Cetus', 48.83, 'Transit', 2017, 5.600, 1.730, 24.7370, true, 'Super-Earth', 235),
('TOI-700 d', 'TOI-700', 'Dorado', 101.40, 'Transit', 2020, 1.720, 1.144, 37.4260, true, 'Terrestrial', 269),
('K2-18 b', 'K2-18', 'Leo', 124.00, 'Transit', 2015, 8.630, 2.610, 32.9396, true, 'Neptune-like', 265),
('HD 209458 b', 'HD 209458', 'Pegasus', 159.00, 'Transit', 1999, 219.000, 15.500, 3.5247, false, 'Hot Jupiter', 1450),
('HR 8799 e', 'HR 8799', 'Pegasus', 133.30, 'Direct Imaging', 2010, 2300.000, 13.100, 18000.0000, false, 'Gas Giant', 1150),
('Kepler-186f', 'Kepler-186', 'Cygnus', 582.00, 'Transit', 2014, 1.400, 1.170, 129.9441, true, 'Terrestrial', 188),
('WASP-76b', 'WASP-76', 'Pisces', 634.00, 'Transit', 2013, 292.000, 20.700, 1.8099, false, 'Hot Jupiter', 2470),
('Kepler-22b', 'Kepler-22', 'Cygnus', 635.00, 'Transit', 2011, 9.100, 2.380, 289.8623, true, 'Super-Earth', 262),
('WASP-12b', 'WASP-12', 'Auriga', 1410.00, 'Transit', 2008, 450.000, 21.300, 1.0914, false, 'Hot Jupiter', 2500),
('Kepler-452b', 'Kepler-452', 'Cygnus', 1799.00, 'Transit', 2015, 5.000, 1.630, 384.8430, true, 'Super-Earth', 265);

-- 4. Deep-Sky Messier Objects
CREATE TABLE messier_catalog (
    messier_id text PRIMARY KEY,
    common_name text NOT NULL,
    object_type text NOT NULL,
    constellation text NOT NULL,
    apparent_magnitude numeric(4,2) NOT NULL,
    distance_light_years integer NOT NULL,
    right_ascension text NOT NULL,
    declination text NOT NULL,
    discoverer text,
    year integer,
    description text NOT NULL
);

INSERT INTO messier_catalog VALUES
('M1', 'Crab Nebula', 'Supernova Remnant', 'Taurus', 8.40, 6500, '05h 34m 32s', '+22° 00′ 52″', 'John Bevis', 1731, 'Remnant of the famous SN 1054 supernova recorded by ancient astronomers; powered by a central pulsar spinning 30 times per second.'),
('M13', 'Great Globular Cluster', 'Globular Cluster', 'Hercules', 5.80, 22200, '16h 41m 41s', '+36° 27′ 37″', 'Edmond Halley', 1714, 'Densely bound sphere of hundreds of thousands of ancient stars; primary target of the 1974 interstellar Arecibo message.'),
('M16', 'Eagle Nebula', 'Diffuse Nebula', 'Serpens', 6.00, 7000, '18h 18m 48s', '-13° 49′ 00″', 'Jean-Philippe de Cheseaux', 1745, 'Vast starburst stellar nursery featuring the towering interstellar dust columns known as the Pillars of Creation.'),
('M27', 'Dumbbell Nebula', 'Planetary Nebula', 'Vulpecula', 7.50, 1360, '19h 59m 36s', '+22° 43′ 16″', 'Charles Messier', 1764, 'The first planetary nebula ever discovered; expanding envelope of luminous gas expelled by a dying asymptotic giant star.'),
('M31', 'Andromeda Galaxy', 'Spiral Galaxy', 'Andromeda', 3.44, 2537000, '00h 42m 44s', '+41° 16′ 09″', 'Abd al-Rahman al-Sufi', 964, 'Nearest major spiral neighbor to the Milky Way, containing over one trillion stars; destined to merge with us in ~4.5 billion years.'),
('M33', 'Triangulum Galaxy', 'Spiral Galaxy', 'Triangulum', 5.72, 2730000, '01h 33m 50s', '+30° 39′ 37″', 'Giovanni Battista Hodierna', 1654, 'Third largest galaxy in our Local Group, containing NGC 604, one of the largest known hydrogen emission starburst regions.'),
('M42', 'Orion Nebula', 'Diffuse Nebula', 'Orion', 4.00, 1344, '05h 35m 17s', '-05° 23′ 28″', 'Nicolas-Claude Fabri de Peiresc', 1610, 'Brilliant naked-eye nebula energized by the compact Trapezium cluster of massive young stars carving a cavity into molecular gas.'),
('M45', 'Pleiades (Seven Sisters)', 'Open Cluster', 'Taurus', 1.60, 444, '03h 47m 24s', '+24° 07′ 00″', 'Known since antiquity', NULL, 'Luminous cluster of hot blue B-type stars illuminated by an interstellar reflection dust cloud they are currently passing through.'),
('M51', 'Whirlpool Galaxy', 'Spiral Galaxy', 'Canes Venatici', 8.40, 23000000, '13h 29m 52s', '+47° 11′ 43″', 'Charles Messier', 1773, 'Archetypal grand-design spiral galaxy with sweeping dust lanes and ongoing starburst activity triggered by interaction with NGC 5195.'),
('M57', 'Ring Nebula', 'Planetary Nebula', 'Lyra', 8.80, 2570, '18h 53m 35s', '+33° 01′ 45″', 'Antoine Darquier de Pellepoix', 1779, 'Classic glowing torus of ionized gas with a central white dwarf star whose progenitor shed its outer envelope ~4,000 years ago.'),
('M81', 'Bode''s Galaxy', 'Spiral Galaxy', 'Ursa Major', 6.94, 11800000, '09h 55m 33s', '+69° 03′ 55″', 'Johann Elert Bode', 1774, 'Bright grand-design spiral galaxy housing a supermassive central black hole of approximately 70 million solar masses.'),
('M82', 'Cigar Galaxy', 'Starburst Galaxy', 'Ursa Major', 8.41, 12000000, '09h 55m 52s', '+69° 40′ 47″', 'Johann Elert Bode', 1774, 'Starburst galaxy generating new stars ten times faster than normal spirals, venting energetic superwinds of hydrogen gas.'),
('M87', 'Virgo A', 'Elliptical Galaxy', 'Virgo', 8.60, 53500000, '12h 30m 49s', '+12° 23′ 28″', 'Charles Messier', 1781, 'Dominant supergiant elliptical galaxy anchored by the first directly imaged black hole (M87*, 6.5 billion solar masses) launching a relativistic jet.'),
('M104', 'Sombrero Galaxy', 'Spiral Galaxy', 'Virgo', 8.00, 31100000, '12h 39m 59s', '-11° 37′ 23″', 'Pierre Méchain', 1781, 'Striking edge-on spiral galaxy distinguished by a prominent dust lane encircling a brilliant bulbous central stellar bulge.');

-- 5. Constellations
CREATE TABLE constellations (
    abbreviation text PRIMARY KEY,
    name text NOT NULL,
    genitive text NOT NULL,
    brightest_star text NOT NULL,
    area_sq_deg numeric(6,1) NOT NULL,
    quadrant text NOT NULL
);

INSERT INTO constellations VALUES
('And', 'Andromeda', 'Andromedae', 'Alpheratz', 722.3, 'NQ1'),
('Aqr', 'Aquarius', 'Aquarii', 'Sadalsuud', 979.8, 'SQ4'),
('Aur', 'Auriga', 'Aurigae', 'Capella', 657.4, 'NQ2'),
('CMa', 'Canis Major', 'Canis Majoris', 'Sirius', 380.1, 'SQ2'),
('CVn', 'Canes Venatici', 'Canum Venaticorum', 'Cor Caroli', 465.2, 'NQ3'),
('Cen', 'Centaurus', 'Centauri', 'Rigil Kentaurus', 1060.4, 'SQ3'),
('Cet', 'Cetus', 'Ceti', 'Diphda', 1231.4, 'SQ1'),
('Cnc', 'Cancer', 'Cancri', 'Tarf', 505.9, 'NQ2'),
('Cru', 'Crux', 'Crucis', 'Acrux', 68.4, 'SQ3'),
('Cyg', 'Cygnus', 'Cygni', 'Deneb', 804.0, 'NQ4'),
('Dor', 'Dorado', 'Doradus', 'Alpha Doradus', 179.2, 'SQ1'),
('Her', 'Hercules', 'Herculis', 'Kornephoros', 1225.1, 'NQ3'),
('Leo', 'Leo', 'Leonis', 'Regulus', 947.0, 'NQ2'),
('Lyr', 'Lyra', 'Lyrae', 'Vega', 286.5, 'NQ4'),
('Oph', 'Ophiuchus', 'Ophiuchi', 'Rasalhague', 948.3, 'SQ3'),
('Ori', 'Orion', 'Orionis', 'Rigel', 594.1, 'NQ1'),
('Peg', 'Pegasus', 'Pegasi', 'Enif', 1121.3, 'NQ4'),
('Psc', 'Pisces', 'Piscium', 'Alpherg', 889.4, 'NQ1'),
('Sco', 'Scorpius', 'Scorpii', 'Antares', 496.8, 'SQ3'),
('Ser', 'Serpens', 'Serpentis', 'Unukalhai', 636.9, 'NQ3'),
('Tau', 'Taurus', 'Tauri', 'Aldebaran', 797.2, 'NQ1'),
('Tri', 'Triangulum', 'Trianguli', 'Beta Trianguli', 131.8, 'NQ1'),
('UMa', 'Ursa Major', 'Ursae Majoris', 'Alioth', 1279.7, 'NQ2'),
('Vir', 'Virgo', 'Virginis', 'Spica', 1294.4, 'SQ3'),
('Vul', 'Vulpecula', 'Vulpeculae', 'Anser', 268.2, 'NQ4');

-- Indexes
CREATE INDEX idx_moons_planet ON moons(planet_name);
CREATE INDEX idx_exoplanets_hz ON exoplanets(habitable_zone);
CREATE INDEX idx_exoplanets_dist ON exoplanets(distance_light_years);
CREATE INDEX idx_messier_type ON messier_catalog(object_type);

-- Explorer read privileges
GRANT USAGE ON SCHEMA public TO explorer;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO explorer;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO explorer;
GRANT CONNECT ON DATABASE astronomy TO explorer;

