// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·scanObjectLargeSVE(SB), NOSPLIT, $0-56
	// R0 = Current address in span
	MOVD mem+0(FP), R0
	// R1 = Curren address in scan buffer
	MOVD bufp+8(FP), R1

  MOVD elemdiff+32(FP), R4
  MOVD limit+40(FP), R5
 
loopArray:
  // R2 pointer to ptrsize
  MOVD ptrsize+16(FP), R2

  // R3 pointer to ptrmap
  MOVD ptrmap+24(FP), R3

	// Align loop to a fetch line so that performance is less sensitive
	// to how this function ends up laid out in memory. This is a hot
	// function in the GC, and this is a tight loop. We don't want
	// performance to waver wildly due to unrelated changes.
	PCALIGN $32
loopElement:

  WORD $0x85800060;// 	ldr	p0, [x3]              
  WORD $0x25208009;// 	cntp	x9, p0, p0.b        
  CBZ R9, skip
  WORD $0x05314004;// 	punpkhi	p4.h, p0.b        
  WORD $0x05304000;// 	punpklo	p0.h, p0.b        
  WORD $0x05314002;// 	punpkhi	p2.h, p0.b        
  WORD $0x05304000;// 	punpklo	p0.h, p0.b        
  WORD $0x05314086;// 	punpkhi	p6.h, p4.b        
  WORD $0x05304084;// 	punpklo	p4.h, p4.b        
  WORD $0x05314001;// 	punpkhi	p1.h, p0.b        
  WORD $0x05304000;// 	punpklo	p0.h, p0.b        
  WORD $0x05314043;// 	punpkhi	p3.h, p2.b        
  WORD $0x05304042;// 	punpklo	p2.h, p2.b        
  WORD $0x05314085;// 	punpkhi	p5.h, p4.b        
  WORD $0x05304084;// 	punpklo	p4.h, p4.b        
  WORD $0x053140c7;// 	punpkhi	p7.h, p6.b        
  WORD $0x053040c6;// 	punpklo	p6.h, p6.b        
  WORD $0x85804000;// 	ldr	z0, [x0]              
  WORD $0x25c08010;// 	cmpne	p0.d, p0/z, z0.d, #0
  WORD $0x05e18000;// 	compact	z0.d, p0, z0.d    
  WORD $0x25e08006;// 	cntp	x6, p0, p0.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  WORD $0x8b060c21;// 	add	x1, x1, x6, lsl #3    
  WORD $0x85804400;// 	ldr	z0, [x0, #1, mul vl]  
  WORD $0x25c08411;// 	cmpne	p1.d, p1/z, z0.d, #0
  WORD $0x05e18400;// 	compact	z0.d, p1, z0.d    
  WORD $0x25e08426;// 	cntp	x6, p1, p1.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  WORD $0x8b060c21;// 	add	x1, x1, x6, lsl #3    
  WORD $0x85804800;// 	ldr	z0, [x0, #2, mul vl]  
  WORD $0x25c08812;// 	cmpne	p2.d, p2/z, z0.d, #0
  WORD $0x05e18800;// 	compact	z0.d, p2, z0.d    
  WORD $0x25e08846;// 	cntp	x6, p2, p2.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  WORD $0x8b060c21;// 	add	x1, x1, x6, lsl #3    
  WORD $0x85804c00;// 	ldr	z0, [x0, #3, mul vl]  
  WORD $0x25c08c13;// 	cmpne	p3.d, p3/z, z0.d, #0
  WORD $0x05e18c00;// 	compact	z0.d, p3, z0.d    
  WORD $0x25e08c66;// 	cntp	x6, p3, p3.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  WORD $0x8b060c21;// 	add	x1, x1, x6, lsl #3    
  WORD $0x85805000;// 	ldr	z0, [x0, #4, mul vl]  
  WORD $0x25c09014;// 	cmpne	p4.d, p4/z, z0.d, #0
  WORD $0x05e19000;// 	compact	z0.d, p4, z0.d    
  WORD $0x25e09086;// 	cntp	x6, p4, p4.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  WORD $0x8b060c21;// 	add	x1, x1, x6, lsl #3    
  WORD $0x85805400;// 	ldr	z0, [x0, #5, mul vl]  
  WORD $0x25c09415;// 	cmpne	p5.d, p5/z, z0.d, #0
  WORD $0x05e19400;// 	compact	z0.d, p5, z0.d    
  WORD $0x25e094a6;// 	cntp	x6, p5, p5.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  WORD $0x8b060c21;// 	add	x1, x1, x6, lsl #3    
  WORD $0x85805800;// 	ldr	z0, [x0, #6, mul vl]  
  WORD $0x25c09816;// 	cmpne	p6.d, p6/z, z0.d, #0
  WORD $0x05e19800;// 	compact	z0.d, p6, z0.d    
  WORD $0x25e098c6;// 	cntp	x6, p6, p6.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  WORD $0x8b060c21;// 	add	x1, x1, x6, lsl #3    
  WORD $0x85805c00;// 	ldr	z0, [x0, #7, mul vl]  
  WORD $0x25c09c17;// 	cmpne	p7.d, p7/z, z0.d, #0
  WORD $0x05e19c00;// 	compact	z0.d, p7, z0.d    
  WORD $0x25e09ce6;// 	cntp	x6, p7, p7.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  WORD $0x8b060c21;// 	add	x1, x1, x6, lsl #3    

skip:
  WORD $0x04225702;// 	addvl	x2, x2, #-8
  WORD $0x04205100;// 	addvl	x0, x0, #8 
  WORD $0x04635023;// 	addpl	x3, x3, #1 
	CMP ZR, R2
	BGT loopElement
  
  ADD R4, R0, R0
  CMP R5, R0
  BLT  loopArray

end: 
	MOVD bufp+8(FP), R0
  SUB R0, R1, R0
  LSR $3, R0, R0
  MOVD R0, count+48(FP)
	RET
