package inc

/************************************************************************

      Copyright 2008 Mark Pictor

  This file is part of RS274NGC.

  RS274NGC is free software: you can redistribute it and/or modify
  it under the terms of the GNU General Public License as published by
  the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  RS274NGC is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
  GNU General Public License for more details.

  You should have received a copy of the GNU General Public License
  along with RS274NGC.  If not, see <http://www.gnu.org/licenses/>.

  This software is based on software that was produced by the National
  Institute of Standards and Technology (NIST).

  ************************************************************************/

/* cxxcam - C++ CAD/CAM driver library.
 * Copyright (C) 2013  Nicholas Gill
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

/*
 * codes.h
 *
 *  Created on: 2013-08-27
 *      Author: nicholas
 */

// G Codes are symbolic to be dialect-independent in source code
type GCodes int

const (
	_      GCodes = iota
	G_0           = 0
	G_1           = 10  /*G01------直线插补                G1 linear interpolation                                  */
	G_2           = 20  /*G02------顺时针方向圆弧插补       G2 circular/helical interpolation (clockwise)           */
	G_3           = 30  /*G03------逆时针方向圆弧插补       G3 circular/helical interpolation (counterclockwise)    */
	G_4           = 40  /*G04------定时暂停                G4 dwell                                                 */
	G_10          = 100 /*G10 coordinate system origin setting*/
	G_17          = 170 /* G17------加工XY平面               G17 XY-plane selection    */
	G_18          = 180 /* G18------加工XZ平面               G18 XZ-plane selection    */
	G_19          = 190 /* G19------加工YZ平面               G19 YZ-plane selection    */
	G_20          = 200 /* G20------子程序调用               G20 inch system selection */
	G_21          = 210 /*G21 millimeter system selection*/
	G_28          = 280 /*G28 return to home*/
	G_30          = 300 /*G30 return to secondary home*/
	G_38_2        = 382 /*G38.2 straight probe*/
	G_40          = 400 /*G40------刀具补偿/刀具偏置注销     G40 cancel cutter radius compensation      */
	G_41          = 410 /*G41------刀具补偿——左             G41 start cutter radius compensation left  */
	G_42          = 420 /*G42------刀具补偿——右             G42 start cutter radius compensation right */
	G_43          = 430 /*G43------刀具偏置——正             G43 tool length offset (plus)              */
	G_49          = 490 /*G49------刀具偏置0/+                G49 cancel tool length offset*/
	G_53          = 530 /*G53------直线偏移，注销            G53 motion in machine coordinate system  */
	G_54          = 540 /*G54------直线偏移x                G54 use preset work coordinate system 1   */
	G_55          = 550 /*G55------直线偏移y                G55 use preset work coordinate system 2   */
	G_56          = 560 /*G56------直线偏移z                G56 use preset work coordinate system 3   */
	G_57          = 570 /*G57------直线偏移xy　             G57 use preset work coordinate system 4   */
	G_58          = 580 /*G58------直线偏移xz               G58 use preset work coordinate system 5   */
	G_59          = 590 /*G59------直线偏移yz               G59 use preset work coordinate system 6   */
	G_59_1        = 591 /*G59.1 use preset work coordinate system 7*/
	G_59_2        = 592 /*G59.2 use preset work coordinate system 8*/
	G_59_3        = 593 /*G59.3 use preset work coordinate system 9*/
	G_61          = 610 /*G61------准确路径方式（中）           G61 set path control mode: exact path*/
	G_61_1        = 611 /*                                    G61.1 set path control mode: exact stop*/
	G_64          = 640 /*G64 set path control mode: continuous*/
	G_80          = 800
	G_81          = 810
	G_82          = 820 /*G82 canned cycle: drilling with dwell*/
	G_83          = 830 /*G83 canned cycle: peck drilling*/
	G_84          = 840 /*G84 canned cycle: right hand tapping*/
	G_85          = 850 /*G85 canned cycle: boring, no dwell, feed out*/
	G_86          = 860 /*G86 canned cycle: boring, spindle stop, rapid out*/
	G_87          = 870 /*G87 canned cycle: back boring*/
	G_88          = 880 /*G88 canned cycle: boring, spindle stop, manual out*/
	G_89          = 890 /*G89 canned cycle: boring, dwell, feed out*/
	G_90          = 900 /*G90------绝对尺寸                  G90 absolute distance mode                         */
	G_91          = 910 /*G91------相对尺寸                  G91 incremental distance mode                      */
	G_92          = 920 /*G92------预制坐标                  G92 offset coordinate systems and set parameters   */
	G_92_1        = 921 /*G92.1 cancel offset coordinate systems and set parameters to zero*/
	G_92_2        = 922 /*G92.2 cancel offset coordinate systems but do not reset parameters*/
	G_92_3        = 923 /*G92.3 apply parameters to offset coordinate systems*/
	G_93          = 930 /*G93------时间倒数，进给率          G93 inverse time feed rate mode     */
	G_94          = 940 /*G94------进给率，每分钟进给        G94 units per minute feed rate mode */
	G_98          = 980
	G_99          = 990
)

/*
G05------通过中间点圆弧插补
G06------抛物线插补
G07------Z 样条曲线插补
G08------进给加速
G09------进给减速
G10------数据设置
G16------极坐标编程

G22------半径尺寸编程方式
G220-----系统操作界面上使用
G23------直径尺寸编程方式
G230-----系统操作界面上使用
G24------子程序结束
G25------跳转加工
G26------循环加工


G30------倍率注销
G31------倍率定义
G32------等螺距螺纹切削，英制
G33------等螺距螺纹切削，公制
G34------增螺距螺纹切削
G35------减螺距螺纹切削
G44------刀具偏置——负
G45------刀具偏置+/+
G46------刀具偏置+/-
G47------刀具偏置-/-
G48------刀具偏置-/+

G50------刀具偏置0/-
G51------刀具偏置+/0
G52------刀具偏置-/0



G60------准确路径方式（精）


G62------准确路径方式（粗）
G63------攻螺纹
G68------刀具偏置，内角
G69------刀具偏置，外角
G70------英制尺寸 寸
G71------公制尺寸 毫米
G74------回参考点(机床零点)
G75------返回编程坐标零点
G76------车螺纹复合循环












G331-----螺纹固定循环



G93------时间倒数，进给率          G93 inverse time feed rate mode
G94------进给率，每分钟进给        G94 units per minute feed rate mode
G95------进给率，每转进给
G96------恒线速度控制
G97------取消恒线速度控制
                                   G98 initial level return in canned cycles
                                   G99 R-point level return in canned cycles


*/
