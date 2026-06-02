'use client'

import React, {JSX, useContext, useState} from "react";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    OptionalArrayOfType,
    OptionalSimpleKey,
    RequiredKey,
    DisplayInput,
    ImportDisplayInput,
    resolvePicsFormData,
    setFormImages,
    OptionalKey,
    HandleJsonResponse,
    setFormData,
    ListPageItems,
    importUrlFor,
    viewUrlFor,
    ImportEntryFormWrapper,
    DisplayFormWrapper,
    FlexedArea,
    FlexedSinglesGroup,
    NewEntryFormWrapper,
    ListTableColumn,
    NewColumn,
    NumberToDateStr,
    ListPageTable,
    ExistingRecentSelector,
    CreatedLinkFor,
    ErrHandler,
    MultipartImportRequest,
    updateApiUrlFor,
    createApiUrlFor,
    DoCreateRequest, DoCreateRequestMultipart, DoUpdateMultipartRequest
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    MostRecentImageDisplay,
    PicsDisplay, SporePrintColorArea, SporePrintDensityArea,
} from "@/app/components/formSubcomponents/commonClient";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {
    IsValidNote, NewEntryNotes,
    Note, NotesFormArea
} from "@/app/components/formSubcomponents/notes";
import ID from "@/app/components/formSubcomponents/id";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {SaleArea} from "@/app/components/saleClient";
import {
    AddCreatedQuadColFunction,
    AllEntries,
    Data,
    OnViewCreatorQuadCol
} from "@/app/components/formSubcomponents/shared";
import EntryLink, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {BaseExternalUrl} from "@/app/components/Constants";
import {redirect} from "next/navigation";
import {AssertSpecies, ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {NewMssForm} from "@/app/components/mssClient";
import {FruitData, FruitSelectorCloseable} from "@/app/components/fruitServer";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {NewSporeSwabForm} from "@/app/components/sporeSwabClient";
import {ACL} from "@/app/components/accessControlServer";
import {MssData} from "@/app/components/mssServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {WriteRfidOvcArea} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {AssertPlate} from "@/app/components/plateClient";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {AgarRecipeData} from "@/app/components/agarRecipeServer";

export function AssertSporePrint(input: any): asserts input is SporePrintData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['parent', 'string'],
        ['subspecies', 'string'],
        ['sale', 'string'],
        ['disposed', 'number'],
        ['color', 'string'],
        ['density', 'string'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    let complexRequiredKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (let [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Plate assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
       ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['pics', IsValidPicWithNotesIncoming], // TODO: ensure ok
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function SporePrintImportDisplay({headerLevel}:ImportDisplayInput) { // TODO: USE ONLY FOR EXISTING SPORE PRINTS!
    const [printDate, setPrintDate] = useState<number>(Date.now())
    const [color, setColor] = useState<string | undefined>()
    const [density, setDensity] = useState<string | undefined>()
    const [notes, setNotes] = useState<Note[]>([])
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [image, setImage] = useState<File | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const importEntry = (e: React.MouseEvent)=>{
        e.preventDefault()
        if(!species){
            setErr("A species must be selected")
            return
        }
        let formData = new FormData()
        let dataObj:any = {
            color: color, // TODO: validate on insert
            density: density, // TODO: validate on insert
            printDate:printDate,
            species:species._id,
            notes:notes,
        }
        subspecies && (dataObj.subspecies = subspecies._id) // TODO: validate on in
        setFormData(formData, dataObj)
        if(image!==undefined){
            formData.set("img",image,"img")
        }

        MultipartImportRequest(formData, "sporePrint", AssertSporePrint, setErr, allCookies(cookies))
        // SendMultipartRequest(importUrlFor("sporePrint"), cookies, formData)
        //     .then(HandleJsonResponse)
        //     .then(newItem=>{
        //         AssertSporePrint(newItem)
        //         redirect(viewUrlFor("sporePrint",newItem._id))
        //     })
        //     .catch(ErrHandler(setErr));
    }
    //no parent because we couldn't possibly know it
        return <ImportEntryFormWrapper entryType={"sporePrint"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <DateArea pre={"Print Date: "} readonly={false} when={Date.now()} updateParent={setPrintDate}/>
            <SporePrintColorArea readonly={false} setColor={setColor} />
            <SporePrintDensityArea readonly={false} setDensity={setColor} />
            <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel}/>
            <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies} headerLevel={headerLevel}/>
            <ImageSelector updateParent={setImage}/>
            <NewEntryNotes setNotes={setNotes} />
            <button className={"greenButton"} onClick={importEntry}>{"Create"}</button>
        </ImportEntryFormWrapper>

}

export default function SporePrintDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel
    }: DisplayInput) {
    try {
        AssertSporePrint(data)
        const [initial, setInitial] = useState(data)
        const initNotes: Data<Note>[] = (data.notes || []).map((n)=>{
            return {data: n,disabled:false}
        })
        const [color, setColor] = useState<string | undefined>()
        const [density, setDensity] = useState<string | undefined>()
        const [pics, setPics] = useState(InitialPicsEntries(data.pics))
        const [sale, setSale] = useState(data.sale)
        const [disposed, setDisposed] = useState(data.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>({existing: initNotes,new:[]})
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(data.acl)
        const updateInitial= (updated: SporePrintData)=>{
            setInitial(updated)
            setColor(updated.color)
            setDensity(updated.density)
            setPics(InitialPicsEntries(updated.pics))
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            setAcl(updated.acl)
        }
        const cookies = useContext(CookiesContext)
        const submit = ()=>{
            // sale disposed, project, pics, notes
            let formData = new FormData()
            let dataObj:any={
                color: color,
                density: density,
                sale:sale,
                disposed:disposed,
                notes: notes,
                acl:MarshalAcl(acl),
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(pics)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Set data on form
                setFormData(formData, dataObj)
                setFormImages(formData, "newPic", newImages)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            DoUpdateMultipartRequest("sporePrint",data._id, formData, AssertSporePrint, allCookies(cookies))
                .then(updateInitial)
                .catch(ErrHandler(setErr))

            // SendMultipartRequest(updateApiUrlFor("sporePrint",data._id), cookies, formData)
            //     .then(HandleJsonResponse)
            //     .then((entry) => {
            //         AssertSporePrint(entry)
            //         updateInitial(entry)
            //     })
            //     .catch(ErrHandler(setErr));
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // TODO: test heavily for all
            // TODO: print transfer to agar?
            // TODO: Chain spore print (do not allow after too long) ---------------------------- TODO!!!!
            {
                txt: "Create Spore Swab",
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewSporeSwabForm printIn={data} onCreate={(item: MssData)=>{ // TODO: switch to handlers{{}} format
                        onCreate([{
                            typeText: "Spore Swab",
                            node: <CreatedLinkFor linkId={item._id} typ={"sporeSwab"}/>,
                        }], false)
                    }}/>
                },
            },
            {
                txt: "Create MultiSpore Syringe",
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewMssForm sporePrintIn={data} handlers={{
                        isTopLevel: false,
                        onCreate: (item: MssData)=>{
                            onCreate([{
                                typeText: "Multispore Syringe",
                                node: <CreatedLinkFor linkId={item._id} typ={"mss"}/>,
                            }], false)
                        }
                    }} />
                },
            },
            WriteRfidOvcArea(initial._id),
            // TODO: TRANSFERS SKIPPING SWABS/SYRINGES?! Probably not...
        ]
        return <DisplayFormWrapper entryType={"sporePrint"}>
            <ErrorDisplay err={err} headerLevel={headerLevel} />
            <ID id={data._id} txt={"Spore Print"} entryType={"sporePrint"}/>
            <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
            <MostRecentImageDisplay data={data.mostRecentImage} headerLevel={headerLevel}/>
            <FlexedArea>
                <FlexedSinglesGroup>
                    <DateArea pre={"Print Date: "} readonly={true} when={data.creationDate}/>
                    <DateArea pre={"Last Updated: "} when={data.lastUpdated} readonly={true}/>
                    <DisposedDisplay readonly={false} disposed={disposed} setDisposedOnParent={setDisposed}/>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                    <div>
                        <div>{"Parent: "}</div>
                        {data.parent?
                            <EntryLink props={{displayId:data.parent,linkId:data.parent,entryType:"fruit",openInNewTab:true}}/>
                            :"Store"}
                    </div>
                </FlexedSinglesGroup>
                <FlexedSinglesGroup>
                    <SporePrintColorArea readonly={true} color={data.color}/>
                    <SporePrintDensityArea readonly={true} density={data.density}/>
                    <SaleArea readonly={false} canCreateSale={true} sale={sale} setSale={setSale} headerLevel={headerLevel}/>
                </FlexedSinglesGroup>
            </FlexedArea>
            {/* TODO: area where we can display all the child MSS of this print? */}
            <PicsDisplay pix={initial.pics || []} updateParent={setPics} readonly={readonly} headerLevel={headerLevel}/>{/* Pics */}
            <NotesFormArea initial={initial.notes} readonly={readonly} updateParent={setNotes}/>
            <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
            </TogglableAreaWithDepth>
            {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                e.stopPropagation();
                submit()
            }}>{"Update"}</button>}
        </DisplayFormWrapper>
    } catch (err) {
        return <div>{"ERROR: Spore print data format incorrect: " + err}</div>
    }
}

// Should only be accessible from a fruit's page
export function NewSporePrintForm( // TODO: currently do not like this one...
    {fruitIn, headerLevel, offset, onCreate}: {
        fruitIn?: FruitData
        headerLevel?: number
        offset?: number
        onCreate:(sp: SporePrintData)=>void
}){
    const [fruit, setFruit] = useState<FruitData | undefined>(fruitIn)
    const [pics, setPics] = useState<NewPicWithNotesForm[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    // Spore prints don't have rfid tags, although they have MainCollectionIDs
    const [err, setErr] = useState<string | undefined>(undefined)
    //const [perms, setPerms] = useState<EntryPerms | undefined>()
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const createEntry = (e: React.MouseEvent)=>{
        e.preventDefault()
        if(!fruit){
            setErr("Fruit must be selected")
            return
        }
        // if both pics and notes are empty, do nothing
        if(pics.length===0){
            setErr("Must at least contain one picture")
            return
        }
        let formData = new FormData()
        let dataObj:any = {
            notes:notes,
            fruitId:fruit._id,
            // No perms means inherit from parents?
        }
        // Pics
        dataObj.pics = pics.map(p=>{return {time:p.time,notes:p.notes.new.map(n => {
                return n.data
            })}})
        // Perms
        setFormData(formData, dataObj)
        for (let i = 0; i < pics.length; i++) {
            let toSend = pics[i]
            if (toSend.img === undefined) {
                setErr("new image " + i + " is undefined")
                return
            }
            const fileName = "newPic" + "-" + i
            formData.set(fileName, toSend.img, fileName)
        }
        DoCreateRequestMultipart("sporePrint", formData, AssertSporePrint, allCookies(cookies))
            .then(onCreate)
            .catch(errHandler)

        // SendMultipartRequest(createApiUrlFor("sporePrint"), cookies, formData)
        //     .then(HandleJsonResponse)
        //     .then((resJson)=>{
        //         AssertSporePrint(resJson)
        //         onCreate(resJson)
        //     })
        //     .catch(ErrHandler(setErr));
    }
    if(fruitIn===undefined){
        // TODO: FRUIT SELECTOR!?
    }

    return <NewEntryFormWrapper entryType={"sporePrint"}>
        <ErrorDisplay err={err} headerLevel={headerLevel} offset={offset}/>
        {fruitIn === undefined && <FruitSelectorCloseable onSelect={setFruit}/>}
        <PicsDisplay pix={[]} readonly={false} updateParent={(ps)=>{setPics(ps.new)}} headerLevel={headerLevel} offset={offset}/>
        <NewEntryNotes setNotes={setNotes} />
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function SporePrintListPageTable({data, onClick, withLink}: ListPageItems<SporePrintData>) {
    let cols: ListTableColumn<SporePrintData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v)=>v.species||""),
        NewColumn("Subspec", v=>v.subspecies||"" ),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SporePrintData)=>{
            return <EntryLinkWrapper props={{entry:v,openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v=>{return new SporePrintData(v)}}/>
}
export function SporePrintSelectorTable({data, onClick}: ListPageItems<SporePrintData>) {
    return <SporePrintListPageTable data={data} onClick={onClick} withLink={true} />
}
export function SporePrintSelector(
    {
        doSelect,
        allowCreate // TODO: del?
    }: {
        doSelect: (val: SporePrintData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: SporePrintData[]):JSX.Element=>{
        return <SporePrintSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"sporePrint"} entryTypes={"sporePrints"} doSelect={doSelect} asserter={AssertSporePrint}
                                   table={table}>
        {/* TODO: ok? allowCreate && <NewSporePrintForm handlers={{onCreate: doSelect,isTopLevel: false}}/>*/}
    </ExistingRecentSelector>
}
