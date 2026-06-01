'use client'

import React, {JSX, useContext, useState} from "react";
import {
    IsValidNote,
    NewEntryNotes,
    Note,
    NotesAreaInline,
    NotesFormArea
} from "@/app/components/formSubcomponents/notes";
import {
    AddCreatedQuadColFunction,
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {AddToTransfers, TransfersOutDisplay} from "@/app/components/transferClient";
import {
    CreatedLinkFor,
    DisplayFormWrapper, DoCreateRequest, DoCreateRequestMultipart, DoUpdateMultipartRequest,
    ErrHandler,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    IsString,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    MultipartImportRequest,
    NewColumn,
    NewEntryFormWrapper,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    resolvePicsFormData,
    setFormData,
    setFormImages,
} from "@/app/components/common";
import {
    ErrorDisplay,
    GensFormDisplay,
    MostRecentImageDisplay,
    NameArea,
    ParentDisplay,
    PicsDisplay,
} from "@/app/components/formSubcomponents/commonClient";
import {FruitData} from "@/app/components/fruitServer";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {NewSporePrintForm} from "@/app/components/sporePrintClient";
import {BaseExternalUrl} from "@/app/components/Constants";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {ReadRFIDButton, WriteRfidOvcArea} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import TestAndValidate from "@/app/components/testing/untested";
import {AclDisplay, IsValidAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {NewSporeSwabForm} from "@/app/components/sporeSwabClient";
import {SporeSwab} from "@/app/components/sporeSwabServer";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {OnViewCreatorsQuadColArea, OvcForXfers} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";

export function AssertFruit(input: any): asserts input is FruitData {
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
            throw new Error('Bag assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }

    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['subspecies', 'string'],
        ['genSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Bag assertion failure: optional key ' + key + ' was not valid');
        }
    }

    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
        ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Fruit assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', IsString],
        ['prints', IsString],
        ['pics', IsValidPicWithNotesIncoming],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Bag assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export default function FruitDisplay(
    {
        id, readonly, data, headerLevel, openSporesInNewTab, allowPrintCreation, isTopLevel
    }: {
        id: string;
        readonly: boolean;
        isTopLevel: boolean;
        data: any;
        headerLevel?: number;
        openSporesInNewTab?: boolean;
        allowPrintCreation?: boolean;
    }) {
    try {
        AssertFruit(data)
        const [initial, setInitial] = useState(data)

        const [pics, setPics] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(initial.notes))
        // Helper states
        const [transfersOut, setTransfersOut] = useState(data.transfersOut || [])
        const [sporePrints, setSporePrints] = useState(data.prints || [])
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: FruitData) => {
            setInitial(updated)
            setPics(InitialPicsEntries(updated.pics))
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            // Helper states
            setTransfersOut(updated.transfersOut || [])
            setSporePrints(updated.prints || [])
            setAcl(updated.acl)
        }
        // TODO: fix?
        // const sporePrintsArea = () => {
        //     return <div>
        //         <div>{"Spore Prints: "}</div>
        //         {(sporePrints.length === 0) &&
        //             <div>{"None"}</div>}
        //         {sporePrints.map(spid => {
        //             const b58id = spid
        //             return <div key={b58id}>
        //                 <EntryLink props={{
        //                     displayedId: b58id,
        //                     linkId: b58id,
        //                     entryType: "sporePrint",
        //                     openInNewTab: openSporesInNewTab
        //                 }}>{spid}</EntryLink>
        //             </div>
        //         })}
        //     </div>
        // }
        const cookies = useContext(CookiesContext)
        const fruitSubmit = () => {
            // disposed, notes, existing pics
            let formData = new FormData()
            let dataObj: any = {
                notes: notes,
                disposed: disposed,
                acl: acl,
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(pics)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Set data on form
                setFormData(formData, dataObj)
                //body.set("data", JSON.stringify(dataObj))
                setFormImages(formData, "newPic", newImages)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            DoUpdateMultipartRequest("fruit",initial._id, formData, AssertFruit, allCookies(cookies))
                .then(updateInitial)
                .catch(ErrHandler(setErr))

            // SendMultipartRequest2(updateApiUrlFor("fruit",initial._id), body)
            //     .then(HandleJsonResponse)
            //     .then((newEntry) => {
            //         AssertFruit(newEntry)
            //         updateInitial(newEntry)
            //     })
            //     .catch(ErrHandler(setErr));
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            // TODO: setTransfersOut on this as needed!
            // TODO: USE THIS!
            // TODO: OvcForXfers on others, or use TransfersOut???
            OvcForXfers(data._id, "fruit", ["plate", "slant", "jar", "stasisTube"], allCookies(cookies), AddToTransfers(setTransfersOut, transfersOut), "Clone/Transfer Fruit"), // TODO: ensure list correct// TODO: OVC for clone to plate (transfer)
            {
                txt: "Create Spore Swab",
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewSporeSwabForm fruitIn={data} onCreate={(item: SporeSwab) => {
                        onCreate([{
                            typeText: "Spore Swab",
                            node: <CreatedLinkFor linkId={item._id} typ={"sporeSwab"}/>,
                        }], false)
                    }}/>
                },
            },
            {
                txt: "Create Spore Print",
                newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                    return <NewSporePrintForm fruitIn={data}
                                              onCreate={(item: SporePrintData) => {
                                                  onCreate([{
                                                      typeText: "Spore Print",
                                                      node: <CreatedLinkFor linkId={item._id} typ={"sporePrint"}/>,
                                                  }], false)
                                              }}/>
                },
            },
            WriteRfidOvcArea(initial._id),
        ]
        return (
            <DisplayFormWrapper entryType={"fruit"}>
                <ErrorDisplay err={err}/>
                <ID txt={"Fruit"} id={data._id} entryType={"fruit"}/>
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
                <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType}
                                       headerLevel={headerLevel}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated}
                                                    readonly={readonly} disposed={initial.disposed}
                                                    setDisposedOnParent={setDisposed}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <GensFormDisplay gensSinceSpore={initial.genSpore} dontDisplayGensFruitOrSpore={true}
                                         headerLevel={headerLevel}/>
                    </FlexedSinglesGroup>
                </FlexedArea>
                <TransfersOutDisplay thisId={initial._id} thisEntryType={"fruit"} transfersOut={transfersOut}
                                     allowNewTransferCreation={false}/>
                <PicsDisplay pix={initial.pics || []} updateParent={setPics} readonly={readonly}/>{/* Pics */}
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    fruitSubmit()
                }}>{"Update"}</button>}
            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Fruit data format incorrect: " + err}</div>
    }
}

export function NewFruitForm(
    {parentId, parentType, headerLevel, readonly, onCreate}: {
        parentId: string,
        parentType: string,
        headerLevel?: number,
        readonly: boolean,
        onCreate: (f: FruitData) => void
    }) {
    if (readonly) {
        return null
    }
    const [harvestDate, setHarvestDate] = useState(Date.now())
    const [pics, setPics] = useState<NewPicWithNotesForm[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    //const [perms, setPerms] = useState<EntryPerms | undefined>() // inherit from parents
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const newFruitSubmit = () => {
        let formData = new FormData()
        let dataObj: any = {
            parentId: parentId,
            parentType: parentType,
            harvestDate: harvestDate,
        }
        if (notes.length > 0) {
            dataObj.notes = notes
        }
        if (pics.length > 0) {
            dataObj.pics = pics.map(p => {
                return {
                    time: p.time, notes: p.notes.new.map(v => {
                        return v.data
                    })
                }
            })
            for (let i = 0; i < pics.length; i++) {
                let imgi = pics[i].img
                if (imgi === undefined) {
                    setErr("new image #" + i + " was not set!")
                    return
                }
                const filePrefix = "newPic" + "-" + i
                formData.set(filePrefix, imgi, filePrefix)
            }
        }
        setFormData(formData, dataObj)
        DoCreateRequestMultipart("fruit", formData, AssertFruit, allCookies(cookies))
            .then(onCreate)
            .catch(errHandler)
    }
    return (
        <NewEntryFormWrapper entryType={"fruit"}>
            <ErrorDisplay err={err} headerLevel={headerLevel}/>
            <DateArea pre={"Harvest Date: "} readonly={false} updateParent={setHarvestDate}/>
            <PicsDisplay pix={[]} updateParent={v => {
                setPics(v.new)
            }} headerLevel={headerLevel} readonly={false}/>
            <NewEntryNotes setNotes={setNotes}/>

            <input type="submit" value="Submit" onClick={newFruitSubmit} onSubmit={(e) => {
                e.preventDefault();
            }}/>
        </NewEntryFormWrapper>
    )
}

export function FruitImportDisplay({headerLevel}: ImportDisplayInput) { // TODO: USE ONLY FOR FRUITS PURCHASED OR FOUND
    const [parentType, setParentType] = useState<string | undefined>(undefined) // TODO: ensure this is everywhere in ts and go
    const [species, setSpecies] = useState<SpeciesData | undefined>(undefined)
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>(undefined)
    const [imageFile, setImageFile] = useState<File | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>(undefined)
    const cookies = useContext(CookiesContext)
    const submitImportFruit = () => { // TODO: rework so we only have the one image, and the one data set
        if (parentType === undefined) {
            setErr("source area must be set!")
            return
        }
        // TODO: FIX!
        if (parentType !== "store" && parentType !== "outside") { // TODO: ENSURE OK ELSEWHERE
            setErr("parentType must be store or outside!")
            return
        }
        if (species === undefined) {
            setErr("Species must be set!")
            return
        }
        let formData = new FormData()
        let dataObj: any = {
            parentType: parentType,
            species: species._id,
            notes: notes,
        }
        subspecies && (dataObj.subspecies = subspecies?._id)
        imageFile && formData.set("img", imageFile, "img")

        MultipartImportRequest(formData, "fruit", AssertFruit, setErr, allCookies(cookies))
    }
    return <ImportEntryFormWrapper entryType={"fruit"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        {/* Required Fields */}
        {/* TODO: ParentType: FOR "store" OR "outside" ONLY!!!!! */}{/* TODO: THIS!*/}
        <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel}/>
        {/* Optional fields*/}
        <ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies}
                                    headerLevel={headerLevel}/>
        <ImageSelector updateParent={setImageFile}/>
        <NewEntryNotes setNotes={setNotes}/>
        {/* SUBMIT AREA */}
        <button className={"bottomButton greenButton"} onClick={submitImportFruit} >{"Import"}</button>
    </ImportEntryFormWrapper>
}

export function CreateCloneArea( // TODO: this vs NewFruitForm
    {
        fruitId, headerLevel, onCloneCreated, readonly,
    }: {
        fruitId: string,
        headerLevel?: number,
        onCloneCreated: (f: FruitData) => void,
        readonly: boolean,
    }) {
    if (readonly) {
        return null
    }
    const [typeTo, setTypeTo] = useState("plate")
    const [idTo, setIdTo] = useState<string | undefined>()
    const [notes, setNotes] = useState<Note[]>([])
    const [err, setErr] = useState<string | undefined>()
    const errHandler = ErrHandler(setErr)
    const cookies = useContext(CookiesContext)
    const handleCreate = () => {
        const body: any = {
            idFrom: fruitId,
            typeFrom: "fruit",
            typeTo: typeTo,
            idTo: idTo,
            notes: notes,
        }
        DoCreateRequest("clone", body, AssertFruit, allCookies(cookies)) // TODO: ensure ok!
            .then(onCloneCreated)
            .catch(errHandler)
    }
    return <div>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <div>
            <div>{"Create Clone:"}</div>
            <div>
                <TestAndValidate todos={["no need for type?"]}>
                    <div>{"TYPE TO:"}</div>
                </TestAndValidate>
                <select className={"tailwindSelector"} value={typeTo} onSelect={e => {
                    setTypeTo(e.currentTarget.value)
                }} onChange={() => {
                }}>
                    {["plate", "jar", "slant"].map((opt, i) => {
                        return <option value={opt} key={i}>{opt}</option>
                    })}
                </select>
            </div>
            <div>
                <TestAndValidate
                    todos={["validate that this is working properly in typing as well as reading from rfid"]}>
                    <NameArea currentName={idTo} setName={setIdTo} headerTxt={"Select ID: "} readonly={false}
                              headerLevel={headerLevel}/>
                </TestAndValidate>
                <ReadRFIDButton handleTagRead={setIdTo}/>
            </div>
        </div>
        <NewEntryNotes setNotes={setNotes}/>
        <button className={"basicButton"} onClick={e => {
            e.preventDefault()
            handleCreate()
        }}>{"Submit new Clone"}</button>
    </div>
}

export function FruitListPageTable({data, onClick, withLink}: ListPageItems<FruitData>) {
    let cols: ListTableColumn<FruitData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Harvest", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Species", v=>v.species ),
        NewColumn("Subspecies", (v)=>v.subspecies || ""),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: FruitData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"fruit",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}
export function FruitSelectorTable({data, onClick}: ListPageItems<FruitData>) {
    return <FruitListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function FruitSelector(
    {
        doSelect,
    }: {
        doSelect: (val: FruitData | undefined) => void,
    }) {
    const table = (items: FruitData[]):JSX.Element=>{
        return <FruitSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"fruit"} entryTypes={"fruits"} doSelect={doSelect} asserter={AssertFruit}
                                   table={table}/>
}