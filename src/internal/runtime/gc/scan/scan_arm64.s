// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

#include "go_asm.h"
#include "textflag.h"

TEXT ·scanSpanPackedSparseSVE(SB), NOSPLIT, $32-44
  MOVD mem+0(FP), R0
  MOVD bufp+8(FP), R1
  MOVD objMarks+16(FP), R2
  MOVD elemsize+24(FP), R3
  MOVD ptrMask+32(FP), R4
  MOVD R2, R7 
  MOVD $64, R6 // Size of objmask 

  WORD $0x2518e3e0;// 	ptrue	p0.b            
  WORD $0x2598e022;// 	ptrue	p2.s, vl1       
  WORD $0x25004241;// 	not	p1.b, p0/z, p2.b  
  WORD $0x05b44043;// 	rev	p3.s, p2.s        
  WORD $0x25f8c005;// 	mov	z5.d, #0          
  WORD $0x25f8c00f;// 	mov	z15.d, #0         
  WORD $0x05e0386e;// 	mov	z14.d, x3         

  // Long arithmetical multiplication to 2^n-1
expandLoop: 
  WORD $0x858040e0;// 	ldr	z0, [x7]                 
  WORD $0x05af7001;// 	trn1	z1.s, z0.s, z15.s      
  WORD $0x05af7402;// 	trn2	z2.s, z0.s, z15.s      
  WORD $0x04613023;// 	mov	z3.d, z1.d               
  WORD $0x04623044;// 	mov	z4.d, z2.d               
  WORD $0x04d381c3;// 	lsl	z3.d, p0/m, z3.d, z14.d  
  WORD $0x04d381c4;// 	lsl	z4.d, p0/m, z4.d, z14.d  
  WORD $0x04e10461;// 	sub	z1.d, z3.d, z1.d         
  WORD $0x04e20482;// 	sub	z2.d, z4.d, z2.d         
  WORD $0x05a27023;// 	trn1	z3.s, z1.s, z2.s       
  WORD $0x05a27424;// 	trn2	z4.s, z1.s, z2.s       
  WORD $0x05ac8c84;// 	splice	z4.s, p3, z4.s, z4.s 
  WORD $0x04980483;// 	orr	z3.s, p1/m, z3.s, z4.s   
  WORD $0x049808a3;// 	orr	z3.s, p2/m, z3.s, z5.s   
  WORD $0x04643085;// 	mov	z5.d, z4.d               
  WORD $0xe58040e3;// 	str	z3, [x7]                 
  WORD $0x04275027;// 	addvl	x7, x7, #1             
  WORD $0x042657e6;// 	addvl	x6, x6, #-1            
  
  CBNZ R6, expandLoop

  MOVD $8192, R6 // Page size
  PCALIGN $16
scanLoop: 
  // Fetch mask of words for this iteration
  MOVD (R2), R5
  // Skip empty frame 
  CBZ R5, skipFull
  // Load mask 
  WORD $0x85800048;// 	ldr	p8, [x2]            
  WORD $0x05284509;// 	zip2	p9.b, p8.b, p8.b  
  WORD $0x05284108;// 	zip1	p8.b, p8.b, p8.b  
  // Skip first half if empty
  CBZW R5, skipHalf 

  WORD $0x85800080;// 	ldr	p0, [x4]              
  WORD $0x25004100;// 	and	p0.b, p0/z, p8.b, p0.b
  WORD $0x05314004;// 	punpkhi	p4.h, p0.b        
  WORD $0x05304000;// 	punpklo	p0.h, p0.b        
  WORD $0x05314002;// 	punpkhi	p2.h, p0.b        
  WORD $0x05304000;// 	punpklo	p0.h, p0.b        
  WORD $0x05314086;// 	punpkhi	p4.h, p4.b        
  WORD $0x05304084;// 	punpklo	p6.h, p4.b
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
  WORD $0x25e08005;// 	cntp	x5, p0, p0.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85804400;// 	ldr	z0, [x0, #1, mul vl]  
  WORD $0x25c08411;// 	cmpne	p1.d, p1/z, z0.d, #0
  WORD $0x05e18400;// 	compact	z0.d, p1, z0.d    
  WORD $0x25e08425;// 	cntp	x5, p1, p1.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85804800;// 	ldr	z0, [x0, #2, mul vl]  
  WORD $0x25c08812;// 	cmpne	p2.d, p2/z, z0.d, #0
  WORD $0x05e18800;// 	compact	z0.d, p2, z0.d    
  WORD $0x25e08845;// 	cntp	x5, p2, p2.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85804c00;// 	ldr	z0, [x0, #3, mul vl]  
  WORD $0x25c08c13;// 	cmpne	p3.d, p3/z, z0.d, #0
  WORD $0x05e18c00;// 	compact	z0.d, p3, z0.d    
  WORD $0x25e08c65;// 	cntp	x5, p3, p3.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85805000;// 	ldr	z0, [x0, #4, mul vl]  
  WORD $0x25c09014;// 	cmpne	p4.d, p4/z, z0.d, #0
  WORD $0x05e19000;// 	compact	z0.d, p4, z0.d    
  WORD $0x25e09085;// 	cntp	x5, p4, p4.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85805400;// 	ldr	z0, [x0, #5, mul vl]  
  WORD $0x25c09415;// 	cmpne	p5.d, p5/z, z0.d, #0
  WORD $0x05e19400;// 	compact	z0.d, p5, z0.d    
  WORD $0x25e094a5;// 	cntp	x5, p5, p5.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85805800;// 	ldr	z0, [x0, #6, mul vl]  
  WORD $0x25c09816;// 	cmpne	p6.d, p6/z, z0.d, #0
  WORD $0x05e19800;// 	compact	z0.d, p6, z0.d    
  WORD $0x25e098c5;// 	cntp	x5, p6, p6.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85805c00;// 	ldr	z0, [x0, #7, mul vl]  
  WORD $0x25c09c17;// 	cmpne	p7.d, p7/z, z0.d, #0
  WORD $0x05e19c00;// 	compact	z0.d, p7, z0.d    
  WORD $0x25e09ce5;// 	cntp	x5, p7, p7.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  PCALIGN $16

skipHalf: 
  WORD $0x85800480;// 	ldr	p0, [x4, #1, mul vl]  
  WORD $0x25004120;// 	and	p0.b, p0/z, p9.b, p0.b
  WORD $0x2520a525;// 	cntp	x5, p9, p9.b        
  // Skip second half if empty
  CBZ R5, skipFull
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

  WORD $0x85814000;// 	ldr	z0, [x0, #8, mul vl]  
  WORD $0x25c08010;// 	cmpne	p0.d, p0/z, z0.d, #0
  WORD $0x05e18000;// 	compact	z0.d, p0, z0.d    
  WORD $0x25e08005;// 	cntp	x5, p0, p0.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85814400;// 	ldr	z0, [x0, #9, mul vl]  
  WORD $0x25c08411;// 	cmpne	p1.d, p1/z, z0.d, #0
  WORD $0x05e18400;// 	compact	z0.d, p1, z0.d    
  WORD $0x25e08425;// 	cntp	x5, p1, p1.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85814800;// 	ldr	z0, [x0, #10, mul vl] 
  WORD $0x25c08812;// 	cmpne	p2.d, p2/z, z0.d, #0
  WORD $0x05e18800;// 	compact	z0.d, p2, z0.d    
  WORD $0x25e08845;// 	cntp	x5, p2, p2.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85814c00;// 	ldr	z0, [x0, #11, mul vl] 
  WORD $0x25c08c13;// 	cmpne	p3.d, p3/z, z0.d, #0
  WORD $0x05e18c00;// 	compact	z0.d, p3, z0.d    
  WORD $0x25e08c65;// 	cntp	x5, p3, p3.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85815000;// 	ldr	z0, [x0, #12, mul vl] 
  WORD $0x25c09014;// 	cmpne	p4.d, p4/z, z0.d, #0
  WORD $0x05e19000;// 	compact	z0.d, p4, z0.d    
  WORD $0x25e09085;// 	cntp	x5, p4, p4.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85815400;// 	ldr	z0, [x0, #13, mul vl] 
  WORD $0x25c09415;// 	cmpne	p5.d, p5/z, z0.d, #0
  WORD $0x05e19400;// 	compact	z0.d, p5, z0.d    
  WORD $0x25e094a5;// 	cntp	x5, p5, p5.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85815800;// 	ldr	z0, [x0, #14, mul vl] 
  WORD $0x25c09816;// 	cmpne	p6.d, p6/z, z0.d, #0
  WORD $0x05e19800;// 	compact	z0.d, p6, z0.d    
  WORD $0x25e098c5;// 	cntp	x5, p6, p6.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1
  WORD $0x85815c00;// 	ldr	z0, [x0, #15, mul vl] 
  WORD $0x25c09c17;// 	cmpne	p7.d, p7/z, z0.d, #0
  WORD $0x05e19c00;// 	compact	z0.d, p7, z0.d    
  WORD $0x25e09ce5;// 	cntp	x5, p7, p7.d        
  WORD $0xe5804020;// 	str	z0, [x1]              
  ADD R5<<3, R1, R1

skipFull:
  WORD $0x04205200;// 	addvl	x0, x0, #16  
  WORD $0x04265606;// 	addvl	x6, x6, #-16 
  WORD $0x04645044;// 	addpl	x4, x4, #2   
  WORD $0x04625022;// 	addpl	x2, x2, #1   
  CBNZ R6, scanLoop
end:
  MOVD bufp+8(FP), R0
  SUB R0, R1, R0
  LSR $3, R0, R0
  MOVW R0, count+40(FP)
	RET
