INSERT INTO categories (id, title, description, sort_order) VALUES
    ('ropita',     'Ropita y accesorios',            'Prendas por talla, pañales, toallitas y accesorios para vestir a Lian.', 1),
    ('cuidado',    'Cuidado, descanso y desarrollo', 'Higiene, baño, descanso seguro, juguetes y básicos del día a día.',      2),
    ('paseo',      'Paseo',                          'Para salir con Lian: cómo llevarlo y llevar todo lo necesario.',          3),
    ('especiales', 'Regalos especiales y aportes',   'Regalos de mayor valor y la opción de un aporte para los gastos de Lian.', 4);

INSERT INTO gifts (id, category_id, name, variant, description, required, sort_order) VALUES
    ('bodies-3-6',    'ropita',     'Bodies de algodón',          '3–6 meses',                'Bodies para uso diario; talla 3–6 meses.',                                              6, 1),
    ('bodies-6-9',    'ropita',     'Bodies de algodón',          '6–9 meses',                'Bodies para uso diario; talla 6–9 meses.',                                              6, 2),
    ('bodies-9-12',   'ropita',     'Bodies de algodón',          '9–12 meses',               'Bodies para uso diario; talla 9–12 meses.',                                             6, 3),
    ('panales-t2',    'ropita',     'Pañales talla 2',            'Talla 2',                  'Pañales desechables; revisar unidades por paquete.',                                    4, 4),
    ('toalla',        'cuidado',    'Toalla con capucha',         NULL,                       'Toalla suave con capucha.',                                                             1, 1),
    ('saco',          'cuidado',    'Saco de dormir',             'Talla según peso',         'Saco de dormir sin peso, adecuado al tamaño del bebé.',                                 1, 2),
    ('juguetes',      'cuidado',    'Juguetes de alto contraste', 'Desde nacimiento',         'Estímulos visuales sencillos para bebé.',                                               3, 3),
    ('biberon',       'cuidado',    'Biberón anticólicos',        'Según preferencia',        'Solo si la familia decide utilizar biberón.',                                           1, 4),
    ('portabebe',     'paseo',      'Portabebé ergonómico',       'Apto desde nacimiento',    'Elegir modelo adecuado para la edad y seguir instrucciones.',                           1, 1),
    ('panalera',      'paseo',      'Pañalera o mochila',         'Desde nacimiento',         'Pañalera o mochila para llevar lo necesario.',                                          1, 2),
    ('coche',         'especiales', 'Coche para bebé',            'Apto desde recién nacido', 'Coche plegable y apropiado para recién nacidos.',                                       1, 1),
    ('corral',        'especiales', 'Corral / cuna portátil',     NULL,                       'Espacio portátil para dormir o descansar, con superficie firme y plana.',               1, 2),
    ('esterilizador', 'especiales', 'Esterilizador',              'Microondas o eléctrico',   'Para artículos de alimentación cuando corresponda; seguir instrucciones del fabricante.', 1, 3);
