-- Anatomy schema and seed data for rock-sql

DROP TABLE IF EXISTS bones CASCADE;
DROP TABLE IF EXISTS joints CASCADE;
DROP TABLE IF EXISTS organ_systems CASCADE;
DROP TABLE IF EXISTS muscles CASCADE;

-- 1. Bones of the Human Skeleton
CREATE TABLE bones (
    name text PRIMARY KEY,
    latin_name text NOT NULL,
    division text NOT NULL,
    region text NOT NULL,
    shape text NOT NULL,
    count_in_body integer NOT NULL,
    articulates_with text NOT NULL,
    description text NOT NULL
);

INSERT INTO bones VALUES
-- Cranium (Axial)
('Frontal Bone', 'Os frontale', 'Axial Skeleton', 'Cranium', 'Flat Bone', 1, 'Parietal bones, Sphenoid, Ethmoid, Nasal bones, Zygomatic bones', 'Forms the forehead, superior rims of the orbital cavities, and anterior cranial fossa.'),
('Parietal Bone', 'Os parietale', 'Axial Skeleton', 'Cranium', 'Flat Bone', 2, 'Frontal bone, Occipital bone, Temporal bone, Sphenoid bone', 'Paired bones forming the superior and lateral roof of the cranium, joined at the midline sagittal suture.'),
('Temporal Bone', 'Os temporale', 'Axial Skeleton', 'Cranium', 'Flat Bone', 2, 'Parietal bone, Occipital bone, Sphenoid, Zygomatic bone, Mandible', 'Houses the middle and inner ear sensory structures; features the zygomatic process, styloid process, and mastoid process.'),
('Occipital Bone', 'Os occipitale', 'Axial Skeleton', 'Cranium', 'Flat Bone', 1, 'Parietal bones, Temporal bones, Sphenoid, Atlas (C1)', 'Forms the posterior and inferior base of the cranium; contains the foramen magnum through which the spinal cord passes.'),
('Sphenoid Bone', 'Os sphenoidale', 'Axial Skeleton', 'Cranium', 'Irregular Bone', 1, 'Frontal, Parietal, Temporal, Occipital, Ethmoid, Vomer, Zygomatic', 'The keystone of the cranial floor; bat-shaped bone featuring the sella turcica, which cradles the pituitary gland.'),
('Ethmoid Bone', 'Os ethmooidale', 'Axial Skeleton', 'Cranium', 'Irregular Bone', 1, 'Frontal bone, Sphenoid, Nasal bones, Vomer', 'Delicate sponge-like bone situated between the orbits; features the cribriform plate with olfactory foramina for the sense of smell.'),

-- Facial Bones (Axial)
('Mandible', 'Mandibula', 'Axial Skeleton', 'Facial', 'Irregular Bone', 1, 'Temporal bones (via Temporomandibular Joint)', 'The lower jawbone; the largest, strongest, and only movable bone of the adult human skull; anchors the lower teeth.'),
('Maxilla', 'Maxilla', 'Axial Skeleton', 'Facial', 'Irregular Bone', 2, 'Frontal, Ethmoid, Nasal, Zygomatic, Palatine, Vomer, Mandible', 'Paired upper jawbones forming the central anterior face, hard palate, floor of the orbits, and anchors upper teeth.'),
('Zygomatic Bone', 'Os zygomaticum', 'Axial Skeleton', 'Facial', 'Irregular Bone', 2, 'Maxilla, Temporal bone, Frontal bone, Sphenoid bone', 'The cheekbones; articulates with the temporal bone to form the zygomatic arch and lateral orbital rim.'),
('Nasal Bone', 'Os nasale', 'Axial Skeleton', 'Facial', 'Flat Bone', 2, 'Frontal bone, Ethmoid bone, Maxilla', 'Small oblong paired bones forming the bridge of the nose, anchored superiorly to the frontal bone.'),
('Vomer', 'Vomer', 'Axial Skeleton', 'Facial', 'Flat Bone', 1, 'Sphenoid, Ethmoid, Palatine bones, Maxillae', 'Thin, plowshare-shaped bone forming the posterior and inferior portion of the nasal septum.'),

-- Auditory & Neck (Axial)
('Malleus (Hammer)', 'Malleus', 'Axial Skeleton', 'Auditory Ossicles', 'Irregular Bone', 2, 'Tympanic membrane, Incus', 'Lateral-most auditory ossicle in the middle ear; transmits vibrations from the eardrum to the incus.'),
('Incus (Anvil)', 'Incus', 'Axial Skeleton', 'Auditory Ossicles', 'Irregular Bone', 2, 'Malleus, Stapes', 'Middle ossicle bridging sound waves from the malleus to the stapes in the middle ear cavity.'),
('Stapes (Stirrup)', 'Stapes', 'Axial Skeleton', 'Auditory Ossicles', 'Irregular Bone', 2, 'Incus, Oval window of cochlea', 'Smallest and lightest bone in the human body (~3 mm); transmits acoustic vibrations into the fluid-filled cochlea.'),
('Hyoid Bone', 'Os hyoideum', 'Axial Skeleton', 'Neck', 'Irregular Bone', 1, 'None (suspended by stylohyoid ligaments and muscles)', 'U-shaped bone in the anterior neck; the only bone in the human skeleton that does not articulate with any other bone; supports tongue muscles and larynx.'),

-- Vertebral Column (Axial)
('Atlas (C1 Vertebra)', 'Atlas', 'Axial Skeleton', 'Vertebral Column', 'Irregular Bone', 1, 'Occipital condyles (C0), Axis (C2)', 'First cervical vertebra; lacks a body and spinous process; forms the atlanto-occipital joint allowing nodding "yes" movements.'),
('Axis (C2 Vertebra)', 'Axis', 'Axial Skeleton', 'Vertebral Column', 'Irregular Bone', 1, 'Atlas (C1), C3 vertebra', 'Second cervical vertebra; features the prominent vertical odontoid process (dens), serving as a pivot around which the head rotates "no".'),
('Cervical Vertebrae (C3–C7)', 'Vertebrae cervicales', 'Axial Skeleton', 'Vertebral Column', 'Irregular Bone', 5, 'Adjacent vertebrae', 'Smallest movable vertebrae; characterized by transverse foramina for the passage of vertebral arteries to the brain.'),
('Thoracic Vertebrae (T1–T12)', 'Vertebrae thoracicae', 'Axial Skeleton', 'Vertebral Column', 'Irregular Bone', 12, 'Adjacent vertebrae, Rib heads and tubercles', 'Intermediate vertebrae with heart-shaped bodies; distinguished by costal facets for articulation with the 12 pairs of ribs.'),
('Lumbar Vertebrae (L1–L5)', 'Vertebrae lumbales', 'Axial Skeleton', 'Vertebral Column', 'Irregular Bone', 5, 'Adjacent vertebrae, Sacrum', 'Largest and stoutest vertebrae with massive kidney-shaped bodies built to support upper body weight and absorb compressive load.'),
('Sacrum', 'Os sacrum', 'Axial Skeleton', 'Vertebral Column', 'Irregular Bone', 1, 'L5 vertebra, Ilium (paired), Coccyx', 'Triangular bone formed by the fusion of five sacral vertebrae (S1–S5); forms the posterior wall of the pelvic cavity and sacroiliac joints.'),
('Coccyx', 'Os coccygis', 'Axial Skeleton', 'Vertebral Column', 'Irregular Bone', 1, 'Sacrum', 'Vestigial tailbone formed by the fusion of 3 to 5 small coccygeal vertebrae; provides attachment for pelvic floor muscles.'),

-- Thoracic Cage (Axial)
('Sternum', 'Sternum', 'Axial Skeleton', 'Thorax', 'Flat Bone', 1, 'Clavicles, Costal cartilages of Ribs 1–7', 'The breastbone; flat T-shaped bone situated anteriorly at the chest midline; comprises the manubrium, body (gladiolus), and xiphoid process.'),
('True Ribs (Pairs 1–7)', 'Costae verae', 'Axial Skeleton', 'Thorax', 'Flat Bone', 14, 'Thoracic vertebrae (T1–T7), Sternum via costal cartilages', 'Seven pairs of ribs that attach directly to the sternum through their own individual costal cartilages.'),
('False Ribs (Pairs 8–10)', 'Costae spuriae', 'Axial Skeleton', 'Thorax', 'Flat Bone', 6, 'Thoracic vertebrae (T8–T10), Costal cartilage of Rib 7', 'Three pairs of ribs whose costal cartilages merge into the cartilage of the rib immediately above rather than reaching the sternum directly.'),
('Floating Ribs (Pairs 11–12)', 'Costae fluctuantes', 'Axial Skeleton', 'Thorax', 'Flat Bone', 4, 'Thoracic vertebrae (T11–T12) only', 'Two pairs of delicate ribs that have no anterior attachment whatsoever, ending freely in the lateral abdominal muscular wall.'),

-- Pectoral Girdle & Upper Limb (Appendicular)
('Clavicle', 'Clavicula', 'Appendicular Skeleton', 'Pectoral Girdle', 'Long Bone', 2, 'Sternum (manubrium), Scapula (acromion)', 'The collarbone; S-shaped strut bone that braces the shoulder away from the chest, transmitting upper limb forces to the axial skeleton.'),
('Scapula', 'Scapula', 'Appendicular Skeleton', 'Pectoral Girdle', 'Flat Bone', 2, 'Clavicle, Humerus', 'The shoulder blade; triangular flat bone on the posterolateral thorax featuring the spine, acromion, coracoid process, and glenoid fossa.'),
('Humerus', 'Humerus', 'Appendicular Skeleton', 'Upper Limb', 'Long Bone', 2, 'Scapula (glenoid cavity), Radius, Ulna', 'The bone of the arm; articulates proximally with the glenoid fossa of the scapula and distally with the radius and ulna at the elbow.'),
('Radius', 'Radius', 'Appendicular Skeleton', 'Upper Limb', 'Long Bone', 2, 'Humerus, Ulna, Carpal bones (Scaphoid, Lunate)', 'Lateral bone of the forearm (thumb side); pivots around the ulna to produce pronation and supination of the hand.'),
('Ulna', 'Ulna', 'Appendicular Skeleton', 'Upper Limb', 'Long Bone', 2, 'Humerus (trochlea), Radius', 'Medial bone of the forearm (pinky side); features the prominent olecranon process forming the bony point of the elbow.'),
('Carpal Bones (Wrist)', 'Ossa carpi', 'Appendicular Skeleton', 'Upper Limb', 'Short Bone', 16, 'Radius, Metacarpals, adjacent carpals', 'Eight small bones per wrist arranged in two rows: Proximal (Scaphoid, Lunate, Triquetrum, Pisiform) and Distal (Trapezium, Trapezoid, Capitate, Hamate).'),
('Metacarpals', 'Ossa metacarpi', 'Appendicular Skeleton', 'Upper Limb', 'Long Bone', 10, 'Distal carpals, Proximal phalanges', 'Five miniature long bones per hand forming the framework of the palm, numbered I (thumb) through V (little finger).'),
('Phalanges of Hand', 'Phalanges manus', 'Appendicular Skeleton', 'Upper Limb', 'Long Bone', 28, 'Metacarpals, adjacent phalanges', 'Fourteen bones per hand: two in the thumb (pollex: proximal, distal) and three in each finger (proximal, middle, distal).'),

-- Pelvic Girdle & Lower Limb (Appendicular)
('Coxal Bone (Hip Bone)', 'Os coxae', 'Appendicular Skeleton', 'Pelvic Girdle', 'Irregular Bone', 2, 'Sacrum, Opposite coxal bone (pubic symphysis), Femur', 'Formed by the embryonic fusion of three bones: the Ilium (superior blade), Ischium (inferior/posterior), and Pubis (anterior); cup-shaped acetabulum.'),
('Femur', 'Femur', 'Appendicular Skeleton', 'Lower Limb', 'Long Bone', 2, 'Coxal bone (acetabulum), Patella, Tibia', 'The thigh bone; the longest, heaviest, and strongest bone in the human body, capable of supporting up to 30 times adult body weight.'),
('Patella', 'Patella', 'Appendicular Skeleton', 'Lower Limb', 'Sesamoid Bone', 2, 'Femur (patellar surface)', 'The kneecap; largest sesamoid bone in the body, embedded within the quadriceps tendon to increase mechanical leverage of knee extension.'),
('Tibia', 'Tibia', 'Appendicular Skeleton', 'Lower Limb', 'Long Bone', 2, 'Femur, Fibula, Talus', 'The shin bone; large weight-bearing medial bone of the lower leg; forms the medial malleolus of the ankle joint.'),
('Fibula', 'Fibula', 'Appendicular Skeleton', 'Lower Limb', 'Long Bone', 2, 'Tibia, Talus', 'Slender lateral bone of the lower leg; non-weight-bearing; serves as muscular anchor and forms the lateral malleolus stabilizer of the ankle.'),
('Tarsal Bones (Ankle)', 'Ossa tarsi', 'Appendicular Skeleton', 'Lower Limb', 'Short Bone', 14, 'Tibia, Fibula, Metatarsals', 'Seven bones per ankle: Talus (bears body weight), Calcaneus (heel bone), Navicular, Medial/Intermediate/Lateral Cuneiforms, and Cuboid.'),
('Metatarsals', 'Ossa metatarsi', 'Appendicular Skeleton', 'Lower Limb', 'Long Bone', 10, 'Tarsals, Proximal phalanges', 'Five long bones per foot supporting the longitudinal and transverse arches of the sole, numbered I (big toe) through V (little toe).'),
('Phalanges of Foot', 'Phalanges pedis', 'Appendicular Skeleton', 'Lower Limb', 'Long Bone', 28, 'Metatarsals, adjacent phalanges', 'Fourteen bones per foot: two in the hallux (big toe) and three in each of the four lesser toes.');

-- 2. Major Joint Articulations
CREATE TABLE joints (
    name text PRIMARY KEY,
    joint_type text NOT NULL,
    functional_class text NOT NULL,
    bones_involved text NOT NULL,
    movements_allowed text NOT NULL,
    description text NOT NULL
);

INSERT INTO joints VALUES
('Glenohumeral Joint (Shoulder)', 'Synovial Ball-and-Socket', 'Diarthrosis (Freely Movable)', 'Scapula (glenoid cavity), Humerus (head)', 'Flexion, Extension, Abduction, Adduction, Circumduction, Internal & External Rotation', 'The most mobile and least stable joint in the human body; deepened by the fibrocartilaginous glenoid labrum and stabilized by the rotator cuff.'),
('Acetabulofemoral Joint (Hip)', 'Synovial Ball-and-Socket', 'Diarthrosis (Freely Movable)', 'Coxal bone (acetabulum), Femur (head)', 'Flexion, Extension, Abduction, Adduction, Internal & External Rotation', 'Deep, highly congruent weight-bearing ball-and-socket joint encased in powerful capsular ligaments (iliofemoral, pubofemoral, ischiofemoral).'),
('Knee Joint (Tibiofemoral)', 'Synovial Modified Hinge / Condyloid', 'Diarthrosis (Freely Movable)', 'Femur (condyles), Tibia (plateau), Patella', 'Flexion, Extension, Slight medial & lateral rotation when flexed', 'The largest and most complex joint in the human body; cushioned by medial and lateral menisci and stabilized by ACL, PCL, MCL, and LCL ligaments.'),
('Elbow Joint (Humeroulnar)', 'Synovial Hinge', 'Diarthrosis (Freely Movable)', 'Humerus (trochlea), Ulna (trochlear notch)', 'Flexion and Extension only', 'Classic uniaxial hinge joint governed by the interlocking bony congruence of the trochlear notch around the humeral trochlea.'),
('Atlanto-Axial Joint', 'Synovial Pivot', 'Diarthrosis (Freely Movable)', 'Atlas (C1), Axis (C2 dens)', 'Rotation ("No" movement of the head)', 'Uniaxial pivot joint where the odontoid process (dens) of C2 rotates within a ring formed by the anterior arch and transverse ligament of C1.'),
('Atlanto-Occipital Joint', 'Synovial Condyloid (Ellipsoid)', 'Diarthrosis (Freely Movable)', 'Occipital condyles, Atlas (C1 superior facets)', 'Flexion, Extension ("Yes" nodding movement)', 'Biaxial joint allowing flexion and extension of the cranium on the cervical spine with limited lateral bending.'),
('Radiocarpal Joint (Wrist)', 'Synovial Condyloid (Ellipsoid)', 'Diarthrosis (Freely Movable)', 'Radius, Scaphoid, Lunate, Triquetrum', 'Flexion, Extension, Radial deviation, Ulnar deviation, Circumduction', 'Biaxial joint connecting the distal radius and fibrocartilaginous articular disc to the proximal row of carpal bones.'),
('Talocrural Joint (Ankle)', 'Synovial Hinge', 'Diarthrosis (Freely Movable)', 'Tibia (plafond & medial malleolus), Fibula (lateral malleolus), Talus (trochlea)', 'Dorsiflexion and Plantarflexion', 'Mortise-and-tenon hinge joint engineered for sagittal plane locomotion and upright stability.'),
('Temporomandibular Joint (TMJ)', 'Synovial Modified Hinge (Bicondylar)', 'Diarthrosis (Freely Movable)', 'Temporal bone (mandibular fossa), Mandible (condyle)', 'Elevation, Depression, Protraction, Retraction, Lateral excursion', 'The only movable joint in the skull; divided into upper and lower synovial cavities by an intervening fibrocartilaginous articular disc.'),
('Cranial Sutures', 'Fibrous Suture', 'Synarthrosis (Immovable)', 'Frontal, Parietal, Occipital, Temporal bones', 'None (rigid immovable interlocking)', 'Synarthrotic fibrous joints (coronal, sagittal, lambdoid, squamous) that lock adjacent cranial bones together with dense fibrous connective tissue.'),
('Intervertebral Discs', 'Cartilaginous Symphysis', 'Amphiarthrosis (Slightly Movable)', 'Adjacent vertebral bodies (C2 through S1)', 'Slight flexion, extension, lateral bending; shock absorption', 'Fibrocartilaginous pads comprising an outer concentric annulus fibrosus and an inner gelatinous nucleus pulposus for shock absorption.');

-- 3. Organ Systems
CREATE TABLE organ_systems (
    name text PRIMARY KEY,
    primary_organs text NOT NULL,
    homeostatic_function text NOT NULL,
    vital_metrics text NOT NULL,
    major_cell_types text NOT NULL,
    clinical_significance text NOT NULL
);

INSERT INTO organ_systems VALUES
('Cardiovascular / Circulatory System', 'Heart, Blood vessels (Arteries, Capillaries, Veins), Blood', 'Transports oxygen, nutrients, hormones, and immune cells to tissues; collects metabolic waste and carbon dioxide for elimination.', 'Resting cardiac output: ~5 L/min; Normal blood pressure: 120/80 mmHg; Total blood volume: ~5 L', 'Cardiomyocytes, Endothelial cells, Erythrocytes, Leukocytes, Thrombocytes', 'Coronary artery disease, hypertension, heart failure, myocardial infarction, arrhythmias.'),
('Respiratory System', 'Nose, Pharynx, Larynx, Trachea, Bronchi, Lungs, Alveoli, Diaphragm', 'Gas exchange between ambient air and circulating blood (O2 uptake and CO2 release); regulates blood acid-base pH balance.', 'Resting respiratory rate: 12–20 breaths/min; Tidal volume: ~500 mL; Total alveolar surface area: ~70 m²', 'Type I & Type II pneumocytes, Ciliated columnar epithelium, Goblet cells', 'Asthma, Chronic Obstructive Pulmonary Disease (COPD), pneumonia, pulmonary embolism.'),
('Nervous System', 'Brain, Spinal cord, Peripheral nerves, Sensory organs', 'Rapid electrical and chemical communication; coordinates sensory perception, motor control, higher cognition, and autonomic homeostatic reflexes.', 'Brain contains ~86 billion neurons; Nerve impulse conduction speed: up to 120 m/s', 'Neurons (sensory, motor, interneurons), Astrocytes, Oligodendrocytes, Microglia, Schwann cells', 'Stroke, Alzheimer''s disease, Parkinson''s disease, multiple sclerosis, epilepsy.'),
('Musculoskeletal System', '206 Bones, 600+ Skeletal muscles, Tendons, Ligaments, Cartilage', 'Provides mechanical structural framework, protection of internal organs, levers for locomotion, mineral storage (calcium/phosphate), and hematopoiesis.', 'Comprises ~40-45% of total body mass; Hematopoiesis generates ~200 billion red cells daily', 'Osteoblasts, Osteocytes, Osteoclasts, Myocytes (skeletal muscle fibers), Chondrocytes', 'Osteoporosis, bone fractures, osteoarthritis, muscular dystrophy, tendonitis.'),
('Digestive System', 'Mouth, Esophagus, Stomach, Small intestine, Large intestine, Liver, Pancreas, Gallbladder', 'Mechanical and biochemical breakdown of ingested food, nutrient absorption, water reabsorption, and solid waste elimination.', 'Length of GI tract: ~9 meters; Small intestine mucosal absorptive area: ~30 m²', 'Enterocytes, Parietal cells (secrete HCl), Chief cells, Hepatocytes, Pancreatic acinar cells', 'Gastroesophageal reflux (GERD), peptic ulcer disease, celiac disease, cirrhosis, inflammatory bowel disease.'),
('Endocrine System', 'Pituitary gland, Thyroid, Parathyroids, Adrenals, Pancreas (islets), Gonads, Pineal', 'Long-term regulatory communication via hormones secreted into the bloodstream to modulate metabolism, growth, reproduction, and stress.', 'Hormones act in minute picomolar/nanomolar concentrations with high receptor specificity', 'Endocrine cells (e.g. Beta cells secreting insulin, Thyroid follicular cells secreting T3/T4)', 'Diabetes mellitus, hypothyroidism, hyperthyroidism, Cushing''s syndrome, Addison''s disease.'),
('Lymphatic & Immune System', 'Spleen, Thymus, Lymph nodes, Tonsils, Lymphatic vessels, Bone marrow', 'Defends against pathogens, malignant cells, and foreign antigens; drains excess interstitial fluid and transports dietary lipids.', 'Filters ~2–3 L of interstitial lymph fluid back into venous circulation daily', 'T lymphocytes (CD4, CD8), B lymphocytes, Plasma cells, Natural Killer cells, Macrophages, Dendritic cells', 'Lymphedema, lymphomas, autoimmune diseases (lupus, rheumatoid arthritis), immunodeficiency.'),
('Urinary / Renal System', 'Kidneys (paired), Ureters, Urinary bladder, Urethra', 'Filters blood plasma to excrete nitrogenous wastes (urea, uric acid, creatinine); regulates electrolyte balance, fluid volume, and blood pressure.', 'Glomerular Filtration Rate (GFR): ~125 mL/min (~180 L/day); Daily urine output: ~1.5 L', 'Podocytes, Mesangial cells, Renal tubular epithelial cells, Transitional urothelium', 'Chronic kidney disease, acute kidney injury, nephrolithiasis (kidney stones), urinary tract infections.'),
('Integumentary System', 'Skin (Epidermis, Dermis, Hypodermis), Hair, Nails, Sebaceous & Sweat glands', 'External protective physical barrier against mechanical trauma, ultraviolet radiation, and dehydration; thermoregulation; Vitamin D synthesis.', 'Largest organ in human body; Surface area: ~1.5–2 m²; Accounts for ~15% of body weight', 'Keratinocytes, Melanocytes, Langerhans cells, Merkel cells, Dermal fibroblasts', 'Melanoma and skin cancers, burns, dermatitis, psoriasis, acne vulgaris.');

-- 4. Prime Mover Muscles
CREATE TABLE muscles (
    name text PRIMARY KEY,
    latin_name text NOT NULL,
    region text NOT NULL,
    primary_action text NOT NULL,
    origin text NOT NULL,
    insertion text NOT NULL,
    innervating_nerve text NOT NULL,
    description text NOT NULL
);

INSERT INTO muscles VALUES
('Biceps Brachii', 'Musculus biceps brachii', 'Upper Arm (Anterior)', 'Flexion of elbow, Powerful supination of forearm', 'Long head: supraglenoid tubercle of scapula; Short head: coracoid process', 'Radial tuberosity and bicipital aponeurosis', 'Musculocutaneous nerve (C5, C6)', 'Prominent two-headed anterior arm muscle; most powerful forearm supinator when the elbow is flexed at 90 degrees.'),
('Triceps Brachii', 'Musculus triceps brachii', 'Upper Arm (Posterior)', 'Extension of elbow joint', 'Long head: infraglenoid tubercle; Lateral head: posterior humerus; Medial head: distal posterior humerus', 'Olecranon process of ulna', 'Radial nerve (C6, C7, C8)', 'Three-headed posterior arm muscle acting as the primary antagonist to the biceps; prime mover of elbow extension.'),
('Deltoid', 'Musculus deltoideus', 'Shoulder', 'Abduction of arm (15° to 90°); anterior fibers flex/medially rotate, posterior fibers extend/laterally rotate', 'Clavicle (lateral 1/3), Acromion, Spine of scapula', 'Deltoid tuberosity of humerus', 'Axillary nerve (C5, C6)', 'Large triangular inverted delta muscle forming the rounded contour of the shoulder; primary arm abductor after the supraspinatus initiates first 15°.'),
('Pectoralis Major', 'Musculus pectoralis major', 'Chest / Thorax', 'Adduction and medial rotation of arm; clavicular head flexes arm', 'Clavicle (medial 1/2), Sternum, Costal cartilages of Ribs 1–6', 'Lateral lip of intertubercular groove of humerus', 'Lateral and medial pectoral nerves (C5–T1)', 'Large fan-shaped chest muscle powering pushing, hugging, and climbing movements.'),
('Latissimus Dorsi', 'Musculus latissimus dorsi', 'Back (Posterior Thorax)', 'Extension, adduction, and medial rotation of the arm ("swimmer''s muscle")', 'Spinous processes T7–L5, Thoracolumbar fascia, Iliac crest, Lower 3–4 ribs', 'Floor of intertubercular groove of humerus', 'Thoracodorsal nerve (C6, C7, C8)', 'The broadest muscle of the back; prime mover for pulling down and rowing movements; elevates the trunk in chin-ups.'),
('Diaphragm', 'Diaphragma', 'Thoracoabdominal Divider', 'Prime mover of inspiration; flattens and increases thoracic cavity volume', 'Xiphoid process, Costal cartilages Ribs 7–12, Lumbar vertebrae bodies L1–L3', 'Central tendon of the diaphragm', 'Phrenic nerve (C3, C4, C5: "keeps the diaphragm alive")', 'Dome-shaped musculotendinous sheet separating thoracic and abdominal cavities; accounts for ~75% of resting respiratory air volume.'),
('Rectus Abdominis', 'Musculus rectus abdominis', 'Abdomen (Anterior)', 'Flexion of lumbar spine, compression of abdominal viscera', 'Pubic crest and pubic symphysis', 'Costal cartilages of Ribs 5–7 and xiphoid process', 'Intercostal nerves (T7–T11) and Subcostal nerve (T12)', 'Paired longitudinal muscle column separated by the linea alba and segmented by horizontal tendinous intersections ("six-pack").'),
('Gluteus Maximus', 'Musculus gluteus maximus', 'Pelvis / Buttock', 'Powerful extension of the hip joint, lateral rotation and abduction', 'Posterior ilium, Sacrum, Coccyx, Sacrotuberous ligament', 'Iliotibial tract and gluteal tuberosity of femur', 'Inferior gluteal nerve (L5, S1, S2)', 'The largest and heaviest single muscle in the human body; vital for rising from a seated position, running, and climbing stairs.'),
('Quadriceps Femoris (Rectus Femoris)', 'Musculus rectus femoris', 'Thigh (Anterior)', 'Extension of knee joint, Flexion of hip joint', 'Anterior inferior iliac spine (AIIS) and superior acetabular rim', 'Patella (base) via quadriceps tendon, continuing as patellar ligament to tibial tuberosity', 'Femoral nerve (L2, L3, L4)', 'The only muscle of the four quadriceps heads that crosses both the hip and knee joints; acts as a powerful knee extender in kicking.'),
('Hamstrings (Biceps Femoris)', 'Musculus biceps femoris', 'Thigh (Posterior)', 'Flexion of knee joint, Extension of hip joint', 'Long head: ischial tuberosity; Short head: linea aspera of femur', 'Head of fibula and lateral condyle of tibia', 'Sciatic nerve (tibial and common fibular divisions, L5–S2)', 'Posterior thigh muscle group prime mover for knee flexion and sprinting; frequent site of athletic strain injury.'),
('Gastrocnemius', 'Musculus gastrocnemius', 'Calf (Posterior Leg)', 'Plantarflexion of the foot, flexion of knee joint', 'Lateral and medial epicondyles of femur', 'Calcaneus via the Calcaneal (Achilles) tendon', 'Tibial nerve (S1, S2)', 'Prominent superficial two-headed calf muscle that powers running, jumping, and push-off during walking locomotion.'),
('Trapezius', 'Musculus trapezius', 'Neck & Upper Back', 'Elevation, retraction, and upward rotation of the scapula; head extension', 'Occipital bone, Nuchal ligament, Spinous processes C7–T12', 'Clavicle (lateral 1/3), Acromion and spine of scapula', 'Accessory nerve (Cranial Nerve XI) and C3–C4 cervical nerves', 'Broad diamond-shaped superficial back muscle spanning the neck, shoulders, and mid-back; stabilizes and moves the pectoral girdle.'),
('Masseter', 'Musculus masseter', 'Head / Jaw', 'Elevation and protraction of mandible (closing jaw with powerful bite force)', 'Zygomatic arch (inferior border and medial surface)', 'Angle and ramus of the mandible', 'Mandibular nerve (V3 division of Trigeminal Nerve CN V)', 'The strongest muscle in the human body relative to its size; capable of generating over 200 pounds of force on the molar teeth.');

-- Indexes & Permissions
CREATE INDEX idx_bones_division ON bones(division);
CREATE INDEX idx_bones_region ON bones(region);
CREATE INDEX idx_bones_shape ON bones(shape);
CREATE INDEX idx_joints_type ON joints(joint_type);
CREATE INDEX idx_muscles_region ON muscles(region);

GRANT USAGE ON SCHEMA public TO explorer;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO explorer;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO explorer;
GRANT CONNECT ON DATABASE anatomy TO explorer;
