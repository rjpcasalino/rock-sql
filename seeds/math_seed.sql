-- Mathematics schema and seed data for rock-sql

DROP TABLE IF EXISTS constants CASCADE;
DROP TABLE IF EXISTS polyhedra CASCADE;
DROP TABLE IF EXISTS number_classes CASCADE;
DROP TABLE IF EXISTS theorems CASCADE;
DROP TABLE IF EXISTS curves CASCADE;

-- 1. Fundamental Constants
CREATE TABLE constants (
    symbol text PRIMARY KEY,
    name text NOT NULL,
    value_approx text NOT NULL,
    domain text NOT NULL,
    definition text NOT NULL,
    first_named_by text,
    year_introduced integer,
    is_transcendental boolean NOT NULL DEFAULT false,
    is_irrational boolean NOT NULL DEFAULT true,
    significance text NOT NULL
);

INSERT INTO constants VALUES
('π', 'Pi (Archimedes'' Constant)', '3.14159265358979323846…', 'Geometry & Analysis', 'Ratio of a Euclidean circle''s circumference to its diameter (C / d)', 'William Jones (symbolized by Euler)', 1706, true, true, 'Ubiquitous in geometry, trigonometry, Fourier transforms, normal distributions, and general relativity field equations.'),
('e', 'Euler''s Number (Napier''s Constant)', '2.71828182845904523536…', 'Calculus & Analysis', 'Base of the natural logarithm; lim(n→∞) (1 + 1/n)^n = ∑(k=0..∞) 1/k!', 'Jacob Bernoulli (studied by Euler)', 1683, true, true, 'The unique real base whose exponential function equals its own derivative: d/dx(e^x) = e^x. Essential for compound growth and differential equations.'),
('φ', 'Golden Ratio (Divine Proportion)', '1.61803398874989484820…', 'Geometry & Number Theory', '(1 + √5) / 2; positive real root of quadratic x^2 - x - 1 = 0', 'Known to ancients (symbolized after Phidias)', NULL, false, true, 'Limiting ratio of consecutive Fibonacci numbers. The "most irrational" number because its continued fraction expansion is purely [1; 1, 1, 1…].'),
('√2', 'Pythagoras'' Constant', '1.41421356237309504880…', 'Geometry & Algebra', 'Length of the hypotenuse of an isosceles right triangle with unit legs; root of x^2 = 2', 'Pythagoreans (Hippasus)', -500, false, true, 'First quantity proven irrational in human history, triggering a foundational crisis in Greek geometry. Defines the aspect ratio of ISO 216 paper (A4, A3).'),
('γ', 'Euler-Mascheroni Constant', '0.57721566490153286060…', 'Analysis & Number Theory', 'Limiting difference between the harmonic series and natural log: lim(n→∞) (∑(1/k) - ln(n))', 'Leonhard Euler', 1735, false, true, 'Appears in gamma function asymptotics, the Riemann zeta function, and prime number distribution. Whether γ is rational or irrational remains an open problem!'),
('α', 'Feigenbaum Constant (Period Doubling)', '4.66920160910299067185…', 'Chaos & Dynamical Systems', 'Universal ratio of distance between successive period-doubling bifurcation intervals', 'Mitchell Feigenbaum', 1975, false, true, 'Universal scaling constant governing the transition from order into deterministic chaos for any smooth 1D map (e.g. the logistic population growth map).'),
('G', 'Catalan''s Constant', '0.91596559417721901505…', 'Combinatorics & Analysis', 'Alternating sum of reciprocals of odd squares: ∑(n=0..∞) (-1)^n / (2n+1)^2', 'Eugène Charles Catalan', 1865, false, true, 'Prominent in combinatorial evaluations, Dirichlet beta function β(2), quantum electrodynamics, and hyperbolic 3-manifold volumes.'),
('ζ(3)', 'Apéry''s Constant', '1.20205690315959428539…', 'Number Theory', 'Sum of the reciprocals of positive cubes: ∑(n=1..∞) 1/n^3', 'Roger Apéry', 1978, false, true, 'Proven irrational by Roger Apéry in 1978. Arises naturally in quantum field theory (electron gyromagnetic ratio) and blackbody photon gas calculations.'),
('c', 'Speed of Light in Vacuum', '299,792,458 m/s (exact)', 'Physics & Relativity', 'Universal invariant speed limit for causality and massless gauge bosons in spacetime', 'Ole Rømer (first measured)', 1676, false, false, 'Sets the conversion scale between mass and energy: E = mc^2. Unifies space and time coordinates into 4-dimensional Minkowski spacetime.');

-- 2. Polyhedra
CREATE TABLE polyhedra (
    name text PRIMARY KEY,
    family text NOT NULL,
    vertices integer NOT NULL,
    edges integer NOT NULL,
    faces integer NOT NULL,
    face_types text NOT NULL,
    schlafli_symbol text NOT NULL,
    symmetry_group text NOT NULL,
    dual_polyhedron text NOT NULL,
    description text NOT NULL
);

INSERT INTO polyhedra VALUES
('Tetrahedron', 'Platonic Solid', 4, 6, 4, '4 Equilateral Triangles', '{3,3}', 'Tetrahedral (T_d)', 'Self-dual', 'The simplest Platonic solid and 3D simplex. Four identical equilateral triangles meet with 3 at each vertex at 60-degree angles.'),
('Cube (Hexahedron)', 'Platonic Solid', 8, 12, 6, '6 Squares', '{4,3}', 'Octahedral (O_h)', 'Octahedron', 'Regular hexahedron with six square faces meeting at right angles. One of only three regular polyhedra that can tile 3D Euclidean space without gaps.'),
('Octahedron', 'Platonic Solid', 6, 12, 8, '8 Equilateral Triangles', '{3,4}', 'Octahedral (O_h)', 'Cube (Hexahedron)', 'Dual partner of the cube. Eight equilateral triangles meet with 4 at each vertex, defining three mutually perpendicular square cross-sections.'),
('Dodecahedron', 'Platonic Solid', 20, 30, 12, '12 Regular Pentagons', '{5,3}', 'Icosahedral (I_h)', 'Icosahedron', 'Composed of 12 regular pentagons with 3 meeting at each vertex. The golden ratio φ is deeply embedded in its vertex coordinates and diagonals.'),
('Icosahedron', 'Platonic Solid', 12, 30, 20, '20 Equilateral Triangles', '{3,5}', 'Icosahedral (I_h)', 'Dodecahedron', 'Twenty equilateral triangular faces with 5 meeting at each vertex. Structural blueprint for geodesic domes and icosahedral viral capsids.'),
('Cuboctahedron', 'Archimedean Solid', 12, 24, 14, '8 Triangles, 6 Squares', 'r{4,3}', 'Octahedral (O_h)', 'Rhombic Dodecahedron', 'Vector Equilibrium (Buckminster Fuller). The distance from its geometric center to all 12 vertices exactly equals its edge length.'),
('Icosidodecahedron', 'Archimedean Solid', 30, 60, 32, '20 Triangles, 12 Pentagons', 'r{5,3}', 'Icosahedral (I_h)', 'Rhombic Triacontahedron', 'Convex quasiregular polyhedron formed by rectifying either a regular dodecahedron or an icosahedron.'),
('Small Stellated Dodecahedron', 'Kepler-Poinsot Solid', 12, 30, 12, '12 Pentagrams (Star Pentagons)', '{5/2,5}', 'Icosahedral (I_h)', 'Great Dodecahedron', 'Regular star polyhedron formed by stellation of the dodecahedron. 12 intersecting star pentagons with 5 meeting at each vertex.'),
('Great Dodecahedron', 'Kepler-Poinsot Solid', 12, 30, 12, '12 Regular Pentagons', '{5,5/2}', 'Icosahedral (I_h)', 'Small Stellated Dodecahedron', 'Regular non-convex star polyhedron composed of 12 intersecting pentagonal faces intersecting one another in 3D space.');

-- 3. Number Classes & Sequences
CREATE TABLE number_classes (
    name text PRIMARY KEY,
    symbol text,
    oeis_id text,
    definition text NOT NULL,
    recurrence_or_formula text NOT NULL,
    first_terms text NOT NULL,
    open_conjecture text,
    description text NOT NULL
);

INSERT INTO number_classes VALUES
('Prime Numbers', 'ℙ', 'A000040', 'Integers greater than 1 with no positive divisors other than 1 and themselves', 'No algebraic closed form; π(x) ~ x / ln(x)', '2, 3, 5, 7, 11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47', 'Riemann Hypothesis, Goldbach Conjecture, Twin Prime Conjecture', 'The fundamental multiplicative building blocks of all integers according to the Fundamental Theorem of Arithmetic.'),
('Fibonacci Numbers', 'F_n', 'A000045', 'Each term is the sum of the preceding two, starting from 0 and 1', 'F_n = F_{n-1} + F_{n-2}; Binet formula: (φ^n - ψ^n) / √5', '0, 1, 1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 144, 233, 377', 'Are there infinitely many Fibonacci primes?', 'Ubiquitous in nature: reflects logarithmic spiral growth in sunflower seed spirals, pinecones, and branching plants.'),
('Perfect Numbers', 'P', 'A000396', 'Positive integers equal to the sum of their proper positive divisors', 'Euclid-Euler: 2^(p-1) * (2^p - 1) for Mersenne prime (2^p - 1)', '6, 28, 496, 8128, 33550336, 8589869056', 'Are there any odd perfect numbers? (Oldest open question in mathematics)', '6 = 1 + 2 + 3; 28 = 1 + 2 + 4 + 7 + 14. Every known perfect number is even and ends in 6 or 28.'),
('Triangular Numbers', 'T_n', 'A000217', 'Number of dots required to form an equilateral triangle of side length n', 'T_n = n * (n + 1) / 2 = ∑(k=1..n) k', '1, 3, 6, 10, 15, 21, 28, 36, 45, 55, 66, 78, 91, 105, 120', 'Gauss Eureka: Every positive integer is the sum of at most 3 triangular numbers', 'Sum of the first n integers. Equal to binomial coefficient (n+1 choose 2), the number of handshakes among n+1 people.'),
('Catalan Numbers', 'C_n', 'A000108', 'Counts balanced parenthesizations, binary trees, and polygon triangulations', 'C_n = (1 / (n + 1)) * (2n choose n) = (2n)! / ((n + 1)! * n!)', '1, 1, 2, 5, 14, 42, 132, 429, 1430, 4862, 16796, 58786', NULL, 'One of the most versatile sequences in combinatorics; counts valid Dyck paths, non-crossing partitions, and rooted binary trees with n+1 leaves.'),
('Mersenne Primes', 'M_p', 'A000668', 'Prime numbers that are one less than a power of two: 2^p - 1', 'M_p = 2^p - 1 verified via the Lucas-Lehmer primality test', '3, 7, 31, 127, 8191, 131071, 524287, 2147483647', 'Are there infinitely many Mersenne primes?', 'Because of the ultra-fast Lucas-Lehmer test, the largest known primes in mathematics are almost exclusively Mersenne primes.'),
('Factorials', 'n!', 'A000142', 'Product of all positive integers less than or equal to n', 'n! = ∏(k=1..n) k; continuous extension via Euler Gamma function Γ(n+1)', '1, 1, 2, 6, 24, 120, 720, 5040, 40320, 362880, 3628800', NULL, 'Counts permutations of n distinct objects. Asymptotically approximated by Stirling''s formula: n! ~ √(2πn) * (n/e)^n.'),
('Collatz Sequences', '3n+1', 'A006577', 'Dynamical orbit: if n is even, n/2; if n is odd, 3n+1', 'T(n) = n/2 if n ≡ 0 (mod 2) else 3n + 1', 'Starting at 27: reaches peak 9232 and takes 111 steps to reach 1', 'Collatz 3n+1 Conjecture (Does every integer eventually reach the 4-2-1 loop?)', 'Simple to state yet completely unproven. Paul Erdős remarked: "Mathematics is not yet ripe enough for such questions."');

-- 4. Landmark Theorems
CREATE TABLE theorems (
    name text PRIMARY KEY,
    field text NOT NULL,
    prover text NOT NULL,
    year_proven text,
    formal_statement text NOT NULL,
    plain_english text NOT NULL,
    significance text NOT NULL
);

INSERT INTO theorems VALUES
('Pythagorean Theorem', 'Geometry', 'Pythagoras / Mesopotamian Geometers', '-500', 'In any right-angled triangle: a^2 + b^2 = c^2', 'The square on the hypotenuse is equal to the sum of the squares on the other two legs.', 'Underpins Euclidean metrics, the Cartesian distance formula, and inner product spaces.'),
('Euler''s Identity', 'Complex Analysis', 'Leonhard Euler', '1748', 'e^(i * π) + 1 = 0', 'The base of logarithms (e), imaginary unit (i), circle ratio (π), multiplicative unity (1), and zero (0) are united in one line.', 'Celebrated as the most beautiful formula in mathematics for combining five fundamental constants with the three basic arithmetic operations.'),
('Fundamental Theorem of Arithmetic', 'Number Theory', 'Euclid / Carl Friedrich Gauss', '1801', 'Every integer n > 1 can be represented uniquely as a product of prime powers: n = p_1^{a_1} * p_2^{a_2} * … * p_k^{a_k}', 'Every whole number greater than 1 is either a prime itself or can be factored into a unique set of prime numbers.', 'Guarantees the unique factorization property of the integers, making primes the fundamental building blocks of arithmetic.'),
('Fundamental Theorem of Calculus', 'Calculus & Analysis', 'Isaac Newton & Gottfried Wilhelm Leibniz', '1675', 'If f is continuous on [a,b] and F''(x) = f(x), then ∫_a^b f(x)dx = F(b) - F(a)', 'Differentiation (slopes/rates) and integration (areas/accumulations) are exact inverse operations of each other.', 'Transformed calculus from disparate geometric puzzles into the foundational engine of physics, astronomy, and modern engineering.'),
('Fermat''s Last Theorem', 'Algebraic Number Theory', 'Pierre de Fermat (1637) / Andrew Wiles (1994)', '1994', 'No three positive integers a, b, c satisfy a^n + b^n = c^n for any integer n > 2', 'While a^2 + b^2 = c^2 has infinite integer solutions, no whole number solutions exist for cubes, fourth powers, or any higher power.', 'Stood unproven for 358 years; Andrew Wiles'' proof revolutionized modern arithmetic geometry by establishing the Modularity Theorem for elliptic curves.'),
('Gödel''s First Incompleteness Theorem', 'Mathematical Logic', 'Kurt Gödel', '1931', 'Any consistent formal system F capable of expressing basic arithmetic contains true statements that cannot be proven within F.', 'In any mathematical rulebook, there will always be truths that the rules themselves are powerless to prove.', 'Dismantled David Hilbert''s program to establish a complete and consistent axiomatic foundation for all of mathematics.'),
('Four Color Theorem', 'Graph Theory & Topology', 'Kenneth Appel & Wolfgang Haken', '1976', 'Any planar graph can be colored using at most four colors such that no two adjacent regions share the same color.', 'You never need more than four distinct colors to color any map on a flat sheet or sphere so that adjacent regions look different.', 'The first major mathematical theorem to rely fundamentally on automated computer computation for its proof verification.'),
('Central Limit Theorem', 'Probability & Statistics', 'Pierre-Simon Laplace & Aleksandr Lyapunov', '1901', 'For independent and identically distributed random variables with finite variance, the normalized sample mean converges to N(0, 1).', 'When you add together many independent random effects, their combined total inevitably takes the shape of a bell curve.', 'Explains the ubiquitous emergence of the Gaussian normal distribution across physics, error analysis, biology, economics, and polling.');

-- 5. Curves
CREATE TABLE curves (
    name text PRIMARY KEY,
    curve_type text NOT NULL,
    cartesian_equation text,
    polar_equation text,
    parametric_equation text,
    discoverer text,
    year integer,
    physical_property text NOT NULL
);

INSERT INTO curves VALUES
('Cardioid', 'Epicycloid (Roulette)', '(x^2 + y^2 - 2ax)^2 = 4a^2(x^2 + y^2)', 'r = 2a(1 - cos θ)', 'x = a(2 cos t - cos 2t), y = a(2 sin t - sin 2t)', 'Johann Castillon', 1741, 'Heart-shaped curve traced by a point on the rim of a circle rolling around another circle of equal radius; caustic of reflection in coffee cups.'),
('Lemniscate of Bernoulli', 'Algebraic Curve', '(x^2 + y^2)^2 = 2a^2(x^2 - y^2)', 'r^2 = 2a^2 cos(2θ)', 'x = (a√2 cos t)/(1 + sin^2 t), y = (a√2 sin t cos t)/(1 + sin^2 t)', 'Jakob Bernoulli', 1694, 'Figure-eight curve; locus of points where the product of distances to two foci is constant. Inspired the infinity symbol ∞.'),
('Cycloid', 'Roulette (Trochoid)', 'x = r*arccos(1 - y/r) - √(2ry - y^2)', NULL, 'x = r(t - sin t), y = r(1 - cos t)', 'Galileo Galilei & Christiaan Huygens', 1599, 'Tautochrone (equal descent time regardless of starting point) and Brachistochrone (curve of fastest descent under gravity between two points).'),
('Golden Spiral', 'Logarithmic Spiral', NULL, 'r = a * φ^(2θ / π)', 'x = a*e^(b*t)*cos(t), y = a*e^(b*t)*sin(t)', 'Renaissance Geometers', 1509, 'Self-similar spiral whose radius expands by the golden ratio φ for every quarter turn; closely models nautilus shells and hurricane shapes.'),
('Astroid', 'Hypocycloid', 'x^(2/3) + y^(2/3) = a^(2/3)', NULL, 'x = a cos^3 t, y = a sin^3 t', 'Ole Rømer', 1674, 'Four-cusped star curve traced by a circle of radius a/4 rolling inside a circle of radius a; envelope of a ladder sliding against a vertical wall.');

-- Permissions & Indexes
CREATE INDEX idx_constants_domain ON constants(domain);
CREATE INDEX idx_polyhedra_family ON polyhedra(family);
CREATE INDEX idx_theorems_field ON theorems(field);

GRANT USAGE ON SCHEMA public TO explorer;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO explorer;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO explorer;
GRANT CONNECT ON DATABASE math TO explorer;
