db.actas_rrv.drop();
db.actas_rrv.createIndex({ acta_id: 1 }, { unique: true });

db.actas_rrv.insertMany([
  // === CONSISTENTE: mismo codigo_acta que PG, votos 0 (iguales) ===
  {
    acta_id: "5070203329004",
    departamento: "Oruro",
    provincia: "Cercado",
    municipio: "Oruro",
    recinto: "Unidad Educativa 6 de Agosto",
    mesa: "M-04",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 0 },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 0 },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 0 },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 0 }
    ],
    votos_nulos: 0, votos_blancos: 0, total_votos: 0,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-15T10:00:00Z")
  },
  {
    acta_id: "5150103603002",
    departamento: "Potosi",
    provincia: "Tomas Frias",
    municipio: "Potosi",
    recinto: "Colegio Nacional Pichincha",
    mesa: "M-02",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 0 },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 0 },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 0 },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 0 }
    ],
    votos_nulos: 0, votos_blancos: 0, total_votos: 0,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-15T10:30:00Z")
  },
  {
    acta_id: "5160103618010",
    departamento: "Potosi",
    provincia: "Rafael Bustillo",
    municipio: "Uncia",
    recinto: "Escuela de Varones Uncia",
    mesa: "M-10",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 0 },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 0 },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 0 },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 0 }
    ],
    votos_nulos: 0, votos_blancos: 0, total_votos: 0,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-15T11:00:00Z")
  },
  // === INCONSISTENTE: mismo codigo_acta que PG, votos DISTINTOS ===
  {
    acta_id: "7010103817016",
    departamento: "Santa Cruz",
    provincia: "Andres Ibanez",
    municipio: "Santa Cruz de la Sierra",
    recinto: "Colegio Santa Ana",
    mesa: "M-16",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 145 },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 198 },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 42  },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 18  }
    ],
    votos_nulos: 12, votos_blancos: 8, total_votos: 423,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-15T09:00:00Z")
  },
  {
    acta_id: "4040102601011",
    departamento: "Chuquisaca",
    provincia: "Oropeza",
    municipio: "Sucre",
    recinto: "Unidad Educativa Junin",
    mesa: "M-11",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 210 },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 87  },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 31  },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 9   }
    ],
    votos_nulos: 15, votos_blancos: 6, total_votos: 358,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-15T09:30:00Z")
  },
  {
    acta_id: "3020101750007",
    departamento: "Cochabamba",
    provincia: "Cercado",
    municipio: "Cochabamba",
    recinto: "Colegio Simon Bolivar",
    mesa: "M-07",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 189 },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 110 },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 55  },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 22  }
    ],
    votos_nulos: 18, votos_blancos: 11, total_votos: 405,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-15T10:00:00Z")
  },
  {
    acta_id: "5010102993005",
    departamento: "Oruro",
    provincia: "Cercado",
    municipio: "Oruro",
    recinto: "Escuela Belen",
    mesa: "M-05",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 302 },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 45  },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 28  },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 5   }
    ],
    votos_nulos: 9, votos_blancos: 14, total_votos: 403,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-15T11:30:00Z")
  },
  // === SOLO_RRV: existen en MongoDB, NO en PostgreSQL ===
  {
    acta_id: "RRV-9990100001001",
    departamento: "La Paz",
    provincia: "Murillo",
    municipio: "El Alto",
    recinto: "Colegio Ayacucho",
    mesa: "M-01",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 180 },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 120 },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 35  },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 15  }
    ],
    votos_nulos: 30, votos_blancos: 20, total_votos: 400,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-14T08:00:00Z")
  },
  {
    acta_id: "RRV-9990100002003",
    departamento: "La Paz",
    provincia: "Murillo",
    municipio: "La Paz",
    recinto: "Colegio San Calixto",
    mesa: "M-03",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 95  },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 245 },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 18  },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 12  }
    ],
    votos_nulos: 8, votos_blancos: 5, total_votos: 383,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-14T09:00:00Z")
  },
  {
    acta_id: "RRV-9990100003007",
    departamento: "Beni",
    provincia: "Cercado",
    municipio: "Trinidad",
    recinto: "Escuela Trinidad Centro",
    mesa: "M-07",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 134 },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 88  },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 62  },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 41  }
    ],
    votos_nulos: 22, votos_blancos: 16, total_votos: 363,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-14T10:00:00Z")
  },
  {
    acta_id: "RRV-9990100004012",
    departamento: "Tarija",
    provincia: "Cercado",
    municipio: "Tarija",
    recinto: "Colegio Nacional Junin",
    mesa: "M-12",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 78  },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 165 },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 44  },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 29  }
    ],
    votos_nulos: 11, votos_blancos: 7, total_votos: 334,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-14T11:00:00Z")
  },
  {
    acta_id: "RRV-9990100005020",
    departamento: "Pando",
    provincia: "Nicolas Suarez",
    municipio: "Cobija",
    recinto: "Unidad Educativa Cobija",
    mesa: "M-20",
    candidatos: [
      { candidato_id: "P1", nombre: "Candidato MAS-IPSP", votos: 56 },
      { candidato_id: "P2", nombre: "Candidato CC",       votos: 43 },
      { candidato_id: "P3", nombre: "Candidato APB",      votos: 31 },
      { candidato_id: "P4", nombre: "Candidato FRI",      votos: 19 }
    ],
    votos_nulos: 6, votos_blancos: 4, total_votos: 159,
    estado: "PROCESADA", fuente: "RRV",
    fecha_procesado: new Date("2025-03-14T12:00:00Z")
  }
]);

print("Seed completado: " + db.actas_rrv.countDocuments() + " actas insertadas");
