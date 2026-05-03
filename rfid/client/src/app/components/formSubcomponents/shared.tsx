"use client"
// TODO: maybe get rid of use client

import {JSX, useState} from "react";
import {CreatedLinkQuadCol, CreatedLinkTriCol} from "@/app/components/pcRunClient";

export type Data<Type> = {
    data: Type,
    disabled: boolean,
}

export type ListResult<Type> = {
    standard?: Type[]
    recent?: Type[]
}

export type SplitAllEntries<Existing, New> = {
    existing: Data<Existing>[],
    new: Data<New>[],
}

export type AllEntries<Type> = SplitAllEntries<Type, Type>

export function NewFromEntries<Type>(ae: AllEntries<Type>) {
    return ae.new.map(e=>{return e.data})
}

export function InitialToAllEntries<Type>(a?: Type[]):AllEntries<Type>{
    return {existing:a?(a.map((v)=>{
            return {data: v,disabled: false}
        })):[], new: []}
}

export interface RevertableAreaProps<Type> {
    readonly?: boolean,
    current?: AllEntries<Type>, // TODO: ensure everywhere is using this properly
    updateParent?: (entries: AllEntries<Type>) => void,
}

export interface AreaProps<Type> {
    readonly?: boolean,
    initialValues?: Data<Type>[],
    updateParent?: (entries: AllEntries<Type>) => void,
    headerLevel?: number
    headerLevelOffset?: number
}

export interface GroupProps<Type> {
    preexisting: boolean,
    readonly: boolean,
    initialEntries?: Data<Type>[],
    updateParent: (n: Data<Type>[])=>void,
    headerLevel?: number
    headerLevelOffset?: number
    blacklist?: Data<Type>[]
}

export function FormListArea<Type>(listGroup: (props: GroupProps<Type>) => JSX.Element){
    return (
        function ({initialValues, readonly, updateParent, headerLevel, headerLevelOffset}: AreaProps<Type>) {
            const [existing, setExisting] = useState(initialValues || [])
            const [newEntries, setNewEntries] = useState<Data<Type>[]>([])
            const [blacklist, setBlacklist] = useState([...existing,...newEntries])
            const updateInternal = (data: AllEntries<Type>) => {
                updateParent && updateParent(data)
                setBlacklist([...data.existing,...data.new])
            }
            const updateExisting = (data: Data<Type>[]) => {
                let newExisting = [...data]
                updateInternal({existing: newExisting, new: newEntries})
                setExisting(newExisting)
            }
            const updateNew = (data: Data<Type>[]) => {
                let newNew = [...data]
                updateInternal({existing: existing, new: newNew})
                setNewEntries(newNew)
            }
            return <div className="formListArea formArea">
                {listGroup({
                    initialEntries: initialValues || [],
                    preexisting: true,
                    readonly: readonly || false,
                    updateParent: updateExisting,
                    headerLevel: headerLevel,
                    headerLevelOffset: headerLevelOffset,
                })}
                {(readonly || initialValues===undefined) ? null : "new"}
                {readonly ? null : listGroup({
                    initialEntries: [],
                    preexisting: false,
                    readonly: false,
                    updateParent: updateNew,
                    headerLevel: headerLevel,
                    headerLevelOffset: headerLevelOffset,
                })}
            </div>
        }
    )
}

export type OnViewCreatorTriCol = {
    txt: string,
    newCreationArea: OnViewCreatorTriColFunction,
}
export type OnViewCreatorTriColFunction = (onCreate: AddCreatedTriColFunction) => JSX.Element
export type AddCreatedTriColFunction = (newLinks: CreatedLinkTriCol[]) => void

export type OnViewCreatorQuadCol = {
    txt: string,
    newCreationArea: OnViewCreatorQuadColFunction,
}
export type OnViewCreatorQuadColFunction = (onCreate: AddCreatedQuadColFunction) => JSX.Element
export type AddCreatedQuadColFunction = (newLinks: CreatedLinkQuadCol[]) => void


